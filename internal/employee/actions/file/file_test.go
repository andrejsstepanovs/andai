package file_test

import (
	"testing"

	"github.com/andrejsstepanovs/andai/internal/employee/actions/file"
	"github.com/stretchr/testify/assert"
)

func TestBuildPromptTextTmpFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		create := func() string {
			fileName, err := file.BuildPromptTextTmpFile("test content")
			assert.NoError(t, err)
			return fileName
		}

		fileName := create()
		assert.NotEmpty(t, fileName)

		content, err := file.GetContents(fileName)
		assert.NoError(t, err)
		assert.Equal(t, "test content", content)
	})
}
