// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func SetupTestArena(t *testing.T, uniqueName string) *Arena {
	dbPath := fmt.Sprintf("%s_test.db", uniqueName)
	os.Remove(filepath.Join(model.BaseDir, dbPath))
	arena, err := NewArena(dbPath)
	assert.Nil(t, err)
	return arena
}

func setupTestArena(t *testing.T) *Arena {
	model.BaseDir = ".."
	return SetupTestArena(t, "field")
}
