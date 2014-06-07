// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the team list.

package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var officialTeamInfoUrl = "https://my.usfirst.org/frc/scoring/index.lasso?page=teamlist"
var officialTeamInfo map[int][]string

// Shows the team list.
func TeamsGetHandler(w http.ResponseWriter, r *http.Request) {
	renderTeams(w, r, false)
}

// Adds teams to the team list.
func TeamsPostHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyTeamList() {
		renderTeams(w, r, true)
		return
	}

	err := r.ParseForm()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	var teamNumbers []int
	for _, teamNumberString := range strings.Split(r.PostFormValue("teamNumbers"), "\r\n") {
		teamNumber, err := strconv.Atoi(teamNumberString)
		if err == nil {
			teamNumbers = append(teamNumbers, teamNumber)
		}
	}

	for _, teamNumber := range teamNumbers {
		team, err := getOfficialTeamInfo(teamNumber)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		err = db.CreateTeam(team)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}
	renderTeams(w, r, false)
}

// Clears the team list.
func TeamsClearHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyTeamList() {
		renderTeams(w, r, true)
		return
	}

	err := db.TruncateTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	renderTeams(w, r, false)
}

// Shows the page to edit a team's fields.
func TeamEditGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := db.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}

	template, err := template.ParseFiles("templates/edit_team.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		*Team
	}{eventSettings, team}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Updates a team's fields.
func TeamEditPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := db.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}

	err = r.ParseForm()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	team.Name = r.PostFormValue("name")
	team.Nickname = r.PostFormValue("nickname")
	team.City = r.PostFormValue("city")
	team.StateProv = r.PostFormValue("stateProv")
	team.Country = r.PostFormValue("country")
	team.RookieYear, _ = strconv.Atoi(r.PostFormValue("rookieYear"))
	team.RobotName = r.PostFormValue("robotName")
	err = db.SaveTeam(team)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	renderTeams(w, r, false)
}

// Removes a team from the team list.
func TeamDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyTeamList() {
		renderTeams(w, r, true)
		return
	}

	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := db.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}
	err = db.DeleteTeam(team)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	renderTeams(w, r, false)
}

func renderTeams(w http.ResponseWriter, r *http.Request, showErrorMessage bool) {
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := template.ParseFiles("templates/teams.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		Teams            []Team
		ShowErrorMessage bool
	}{eventSettings, teams, showErrorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Returns true if it is safe to change the team list (i.e. no matches/results exist yet).
func canModifyTeamList() bool {
	matches, err := db.GetMatchesByType("qualification")
	if err != nil || len(matches) > 0 {
		return false
	}
	return true
}

// Returns the data for the given team number.
func getOfficialTeamInfo(teamId int) (*Team, error) {
	if officialTeamInfo == nil {
		// Download all team info from the FIRST website if it is not cached.
		resp, err := http.Get(officialTeamInfoUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		re := regexp.MustCompile("(?s).*<PRE>(.*)</PRE>.*")
		teamsCsv := re.FindStringSubmatch(string(body))[1]

		reader := csv.NewReader(strings.NewReader(teamsCsv))
		reader.Comma = '\t'
		reader.FieldsPerRecord = -1
		officialTeamInfo = make(map[int][]string)
		reader.Read() // Ignore header line.
		for {
			fields, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			teamNumber, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}
			officialTeamInfo[teamNumber] = fields
		}
	}

	teamData, ok := officialTeamInfo[teamId]
	var team Team
	if ok {
		rookieYear, _ := strconv.Atoi(teamData[8])
		team = Team{teamId, html.UnescapeString(teamData[2]), html.UnescapeString(teamData[7]),
			html.UnescapeString(teamData[4]), html.UnescapeString(teamData[5]), html.UnescapeString(teamData[6]),
			rookieYear, html.UnescapeString(teamData[9])}
	} else {
		// If no team data exists, just fill in the team number.
		team = Team{Id: teamId}
	}
	return &team, nil
}