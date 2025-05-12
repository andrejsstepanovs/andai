package issue

import (
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newCreateCommand(deps internal.DependenciesLoader) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Creates new Issue. First param Issue Type, Second param Issue Subject, Third param Issue Description",
		RunE: func(_ *cobra.Command, args []string) error {
			d := deps()
			log.Println("Creating new Issue")

			if len(args) < 3 {
				log.Println("Not enough arguments")
				return nil
			}

			projects, err := d.Model.APIGetProjects()
			if err != nil {
				log.Println("Failed to get projects")
				return err
			}
			if len(projects) == 0 {
				log.Println("No projects found")
				return err
			}
			if len(projects) > 1 {
				log.Println("Too many projects found")
				return err
			}
			project := projects[0]

			issueType := args[0]
			issueSubject := args[1]
			issueDescription := args[2]
			log.Printf("Issue Type: %s, Issue Subject: %s, Issue Description: %s", issueType, issueSubject, issueDescription)

			trackerID, err := d.Model.DBGetTrackersByName(issueType)
			if err != nil {
				log.Printf("Failed to find issue type (tracker): %s", issueType)
				return err
			}

			issueDescription = strings.ReplaceAll(issueDescription, "\n", "<br/>")
			issue := redmine.Issue{
				Subject:     issueSubject,
				Description: issueDescription,

				ProjectId: project.Id,
				Project:   &redmine.IdName{Id: project.Id},
				TrackerId: trackerID,
			}

			issue, err = d.Model.CreateIssue(issue)
			if err != nil {
				log.Println("Failed to create issue")
				return err
			}

			log.Printf("Issue created: %d", issue.Id)
			log.Printf("http://localhost:10083/issues/%d", issue.Id)

			return nil
		},
	}
}
