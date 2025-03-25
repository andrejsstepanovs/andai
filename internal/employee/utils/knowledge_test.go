package utils_test

import (
	"testing"

	"github.com/andrejsstepanovs/andai/internal/employee/utils"
	"github.com/stretchr/testify/assert"
)

func TestTabContent(t *testing.T) {
	t.Run("success 1", func(t *testing.T) {
		k := utils.Knowledge{}
		resp := k.TagContent("apple", "Content Line 1\nContent Line 2\nContent Line 3", 3)
		expected := "<apple>\n" +
			"\t\t\tContent Line 1\n" +
			"\t\t\tContent Line 2\n" +
			"\t\t\tContent Line 3\n" +
			"</apple>"

		assert.Equal(t, expected, resp)
	})

	t.Run("success 2", func(t *testing.T) {
		k := utils.Knowledge{}
		resp := k.TagContent("apple_banana", "Content Line 1\nContent Line 2\n\n", 1)
		expected := "<apple_banana>\n" +
			"\tContent Line 1\n" +
			"\tContent Line 2\n" +
			"\t\n" +
			"\t\n" +
			"</apple_banana>"

		assert.Equal(t, expected, resp)
	})

	t.Run("success 3 empty", func(t *testing.T) {
		k := utils.Knowledge{}
		resp := k.TagContent("empty", "", 1)
		expected := "<empty>\n" +
			"\t\n" +
			"</empty>"

		assert.Equal(t, expected, resp)
	})
}
