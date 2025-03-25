package models_test

import (
	"os"
	"testing"

	"github.com/andrejsstepanovs/andai/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_Load_ValidConfig(t *testing.T) {
	curDir, _ := os.Getwd()
	settings, err := models.NewConfig("project", curDir+"/testdata").Load()
	assert.NoError(t, err)
	assert.NotNil(t, settings)
}

func Test_GetSettings_EmptyYAML(t *testing.T) {
	curDir, _ := os.Getwd()
	_, err := models.NewConfig("empty", curDir+"/testdata").Load()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "settings validation err")
}

func Test_GetSettings_MalformedYAML(t *testing.T) {
	curDir, _ := os.Getwd()
	_, err := models.NewConfig("malformed", curDir+"/testdata").Load()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unmarshal errors")
}
