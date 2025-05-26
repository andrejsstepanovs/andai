package knowledge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitSHA1(t *testing.T) {
	content := `	<comment_2 at="2025-05-26 12:33:19">
	### Branch [AI-220](/projects/projectX/repository/bobik2?rev=AI-220)
	1. Commit [4036d72f3120a22abbcc7736daf3bc2c5d05e155](/projects/projectX/repository/projectX/revisions/4036d72f3120a22abbcc7736daf3bc2c5d05e155/diff) - code changes
	</comment_2>
	<comment_3 at="2025-05-26 12:36:18">
	internal/client.go
	</comment_3>
</comments>
`

	expected := []string{"4036d72f3120a22abbcc7736daf3bc2c5d05e155"}
	commits := GitSHA1(content)
	assert.Equal(t, expected, commits)
}
