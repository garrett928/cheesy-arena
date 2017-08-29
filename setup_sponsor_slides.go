// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
//
// Web routes for managing sponsor slides.

package main

import (
	"github.com/Team254/cheesy-arena/model"
	"html/template"
	"net/http"
	"strconv"
)

// Shows the sponsor slides configuration page.
func (web *Web) sponsorSlidesGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := template.ParseFiles("templates/setup_sponsor_slides.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	sponsorSlides, err := web.arena.Database.GetAllSponsorSlides()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		SponsorSlides []model.SponsorSlide
	}{web.arena.EventSettings, sponsorSlides}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Saves the new or modified sponsor slides to the database.
func (web *Web) sponsorSlidesPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	sponsorSlideId, _ := strconv.Atoi(r.PostFormValue("id"))
	sponsorSlide, err := web.arena.Database.GetSponsorSlideById(sponsorSlideId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if r.PostFormValue("action") == "delete" {
		err := web.arena.Database.DeleteSponsorSlide(sponsorSlide)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	} else {
		displayTimeSec, _ := strconv.Atoi(r.PostFormValue("displayTimeSec"))
		if sponsorSlide == nil {
			sponsorSlide = &model.SponsorSlide{Subtitle: r.PostFormValue("subtitle"),
				Line1: r.PostFormValue("line1"), Line2: r.PostFormValue("line2"),
				Image: r.PostFormValue("image"), DisplayTimeSec: displayTimeSec}
			err = web.arena.Database.CreateSponsorSlide(sponsorSlide)
		} else {
			sponsorSlide.Subtitle = r.PostFormValue("subtitle")
			sponsorSlide.Line1 = r.PostFormValue("line1")
			sponsorSlide.Line2 = r.PostFormValue("line2")
			sponsorSlide.Image = r.PostFormValue("image")
			sponsorSlide.DisplayTimeSec = displayTimeSec
			err = web.arena.Database.SaveSponsorSlide(sponsorSlide)
		}
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}

	http.Redirect(w, r, "/setup/sponsor_slides", 302)
}
