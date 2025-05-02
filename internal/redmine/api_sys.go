package redmine

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

// APISyncRepo re-syncs the repository with Redmine
// you can use an HTTP GET submission to automatically refresh Redmine after you committed your modification in your repository
// https://www.redmine.org/projects/redmine/wiki/HowTo_setup_automatic_refresh_of_repositories_in_Redmine_on_commit
func (c *Model) APISyncRepo(project redmine.Project) error {
	baseURL := viper.GetString("redmine.url")
	key := viper.GetString("redmine.api_key")
	if baseURL == "" || key == "" {
		return fmt.Errorf("redmine url or key is empty")
	}

	parsedURL, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return fmt.Errorf("invalid redmine URL: %v", err)
	}
	query := url.Values{}
	query.Add("id", project.Identifier)
	query.Add("key", key)

	parsedURL.Path = path.Join(parsedURL.Path, "sys/fetch_changesets")
	parsedURL.RawQuery = query.Encode()

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("API error: %d %s\nResponse: %s", resp.StatusCode, resp.Status, string(body))
	}

	return nil
}
