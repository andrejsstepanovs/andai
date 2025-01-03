package redmine

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
)

const (
	RelationBlocks = "blocks"
)

func (c *Model) APIGetWorkableIssue(priorities models.Priorities) (redmine.Issue, error) {
	projects, err := c.ApiGetProjects()
	if err != nil {
		return redmine.Issue{}, fmt.Errorf("error redmine issue status: %v", err)
	}

	for _, project := range projects {
		projectIssues, err := c.Api().IssuesOf(project.Id)
		if err != nil {
			return redmine.Issue{}, fmt.Errorf("error redmine issues of project: %v", err)
		}

		dependencies, err := c.issueDependencies(projectIssues)
		if err != nil {
			return redmine.Issue{}, fmt.Errorf("error redmine issue dependencies: %v", err)
		}

		for issueID, blockedByIDs := range dependencies {
			if len(blockedByIDs) == 0 {
				log.Printf("ISSUE: %d is not blocked at all\n", issueID)
				continue
			}
			blocked := make([]string, 0)
			for _, blockedBy := range blockedByIDs {
				blocked = append(blocked, strconv.Itoa(blockedBy))
			}
			log.Printf("ISSUE: %d BLOCKED BY: %v\n", issueID, strings.Join(blocked, ", "))
		}

		//issues := c.getCorrectIssue(projectIssues, priorities)
		//for _, issue := range issues {
		//	for issueID, blockedByIDs := range dependencies {
		//		if issue.Id == issueID {
		//			if len()
		//			fmt.Printf("SKIP ISSUE: %d (%d -> %d)", issue.Id, from, to)
		//			continue
		//		}
		//	}
		//
		//	fmt.Println("WORKABLE ISSUE:", issue.Id)
		//}
	}

	return redmine.Issue{}, nil
}

func (c *Model) issueDependencies(projectIssues []redmine.Issue) (map[int][]int, error) {
	dependencies := make(map[int][]int)
	for _, issue := range projectIssues {
		dependencies[issue.Id] = make([]int, 0)
		relations, err := c.Api().IssueRelations(issue.Id)
		if err != nil && err.Error() != "Not Found" {
			return dependencies, err
		}
		//fmt.Println("RELATIONS:", len(relations))
		for _, relation := range relations {
			if relation.RelationType != RelationBlocks {
				continue
			}
			//fmt.Printf("ISSUE (%d) - %d IS BLOCKED BY %d <- needs to be done first\n", issue.Id, relation.IssueToId, relation.IssueId)
			dependencies[relation.IssueToId] = append(dependencies[relation.IssueId], relation.IssueId)
		}
	}

	return dependencies, nil
}

func (c *Model) getCorrectIssue(issues []redmine.Issue, priorities models.Priorities) []redmine.Issue {
	valid := make([]redmine.Issue, 0)
	for _, priority := range priorities {
		//fmt.Printf("PRIORITY: %q @ %q\n", priority.Type, priority.State)
		for _, issue := range issues {
			if issue.Tracker.Name != priority.Type {
				//fmt.Printf("SKIP %q (%d) - not %q\n", issue.Tracker.Name, issue.Id, priority.Type)
				continue
			}
			if issue.Status.Name != priority.State {
				//fmt.Printf("SKIP %q (%d) - %q != %q\n", issue.Tracker.Name, issue.Id, issue.Status.Name, priority.State)
				continue
			}
			//fmt.Println("ISSUE:", issue.Tracker.Name, issue.Id)
			valid = append(valid, issue)
		}
	}

	return valid
}
