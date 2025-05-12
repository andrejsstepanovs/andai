package settings_test

import (
	"os"
	"testing"

	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/stretchr/testify/assert"
)

func Test_Load_ValidConfig(t *testing.T) {
	curDir, _ := os.Getwd()
	os.Setenv("PROJECT", "project")
	settings, err := settings.NewConfig(curDir + "/testdata").Load()
	assert.NoError(t, err)
	assert.NotNil(t, settings)
}

func Test_GetSettings_EmptyYAML(t *testing.T) {
	curDir, _ := os.Getwd()
	os.Setenv("PROJECT", "empty")
	_, err := settings.NewConfig(curDir + "/testdata").Load()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "settings validation err")
}

func Test_GetSettings_MalformedYAML(t *testing.T) {
	curDir, _ := os.Getwd()
	os.Setenv("PROJECT", "malformed")
	_, err := settings.NewConfig(curDir + "/testdata").Load()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unmarshal errors")
}
