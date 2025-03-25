package utils_test

import (
	"testing"

	"github.com/andrejsstepanovs/andai/internal/employee/utils"
	"github.com/stretchr/testify/assert"
)

func TestBuildPromptTextTmpFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		create := func() string {
			file, err := utils.BuildPromptTextTmpFile("test content")
			assert.NoError(t, err)
			return file
		}

		file := create()
		assert.NotEmpty(t, file)

		content, err := utils.GetFileContents(file)
		assert.NoError(t, err)
		assert.Equal(t, "test content", content)
	})
}
