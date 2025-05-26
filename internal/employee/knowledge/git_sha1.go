package knowledge

import (
	"regexp"
	"strings"
)

func GitSHA1(content string) []string {
	uniqueCommits := make(map[string]bool)

	content = strings.ReplaceAll(content, "[", " ")
	content = strings.ReplaceAll(content, "]", " ")
	content = strings.ReplaceAll(content, "(", " ")
	content = strings.ReplaceAll(content, ")", " ")
	content = strings.ReplaceAll(content, ".", " ")
	content = strings.ReplaceAll(content, ",", " ")
	content = strings.ReplaceAll(content, "`", " ")
	content = strings.ReplaceAll(content, "/", " ")

	sha1Regex := regexp.MustCompile(`\b([a-f0-9]{40})\b`)
	for _, word := range strings.Fields(content) {
		if match := sha1Regex.FindString(word); match != "" {
			commitSha := strings.TrimSpace(match)
			uniqueCommits[commitSha] = true
		}
	}

	commits := make([]string, 0)
	for commit := range uniqueCommits {
		if len(commit) != 40 {
			continue
		}
		commits = append(commits, commit)
	}

	return commits
}
