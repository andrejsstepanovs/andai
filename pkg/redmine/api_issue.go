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

func (c *Model) APIGetWorkableIssues(workflow models.Workflow) ([]redmine.Issue, error) {
	projects, err := c.ApiGetProjects()
	if err != nil {
		return nil, fmt.Errorf("error redmine issue status: %v", err)
	}

	for _, project := range projects {
		projectIssues, err := c.Api().IssuesOf(project.Id)
		if err != nil {
			return nil, fmt.Errorf("error redmine issues of project: %v", err)
		}

		dependencies, err := c.issueDependencies(projectIssues)
		if err != nil {
			return nil, fmt.Errorf("error redmine issue dependencies: %v", err)
		}

		//for issueID, blockedByIDs := range dependencies {
		//	if len(blockedByIDs) == 0 {
		//		log.Printf("Issue (%d) is not blocked at all\n", issueID)
		//		continue
		//	}
		//	blocked := make([]string, 0)
		//	for _, blockedBy := range blockedByIDs {
		//		blocked = append(blocked, strconv.Itoa(blockedBy))
		//	}
		//	log.Printf("Issue (%d) BLOCKED BY: %v\n", issueID, strings.Join(blocked, ", "))
		//}

		cleanedDependencies := c.removeClosedDependencies(dependencies, projectIssues)
		for issueID, blockedByIDs := range cleanedDependencies {
			if len(blockedByIDs) == 0 {
				log.Printf("Issue (%d) is not blocked at all\n", issueID)
				continue
			}
			blocked := make([]string, 0)
			for _, blockedBy := range blockedByIDs {
				blocked = append(blocked, strconv.Itoa(blockedBy))
			}
			log.Printf("Issue (%d) BLOCKED BY: %v\n", issueID, strings.Join(blocked, ", "))
		}

		unblockedIDs := make([]int, 0)
		for issueID, depIDs := range cleanedDependencies {
			if len(depIDs) == 0 {
				unblockedIDs = append(unblockedIDs, issueID)
			}
		}

		if len(unblockedIDs) == 0 {
			log.Println("No workable issues")
			return nil, nil
		}
		ids := make([]string, 0)
		for _, id := range unblockedIDs {
			ids = append(ids, fmt.Sprintf("%d", id))
		}
		fmt.Printf("UNBLOCKED ISSUES (%d) %s\n", len(unblockedIDs), strings.Join(ids, ", "))

		validIssues := make([]redmine.Issue, 0)
		for _, okIssueID := range unblockedIDs {
			for _, issue := range projectIssues {
				if issue.Id == okIssueID {
					validIssues = append(validIssues, issue)
					break
				}
			}
		}

		fmt.Printf("VALID ISSUES TO CHECK PRIORITIES AGAINST (%d)\n", len(validIssues))
		for _, issue := range validIssues {
			fmt.Println("ISSUE:", issue.Tracker.Name, issue.Id)
		}

		issues := c.getCorrectIssue(validIssues, workflow.Priorities, workflow.States)

		return issues, nil
	}

	return nil, nil
}

func (c *Model) removeClosedDependencies(dependencies map[int][]int, issues []redmine.Issue) map[int][]int {
	cleaned := make(map[int][]int)
	for issueID, blockedByIDs := range dependencies {
		if len(blockedByIDs) == 0 {
			cleaned[issueID] = blockedByIDs
			continue
		}
		cleaned[issueID] = make([]int, 0)

		for _, blockedBy := range blockedByIDs {
			// if dont exist then its closed
			isClosed := true
			for _, issue := range issues {
				if issue.Id == blockedBy {
					isClosed = false
					break
				}
			}

			if !isClosed {
				cleaned[issueID] = append(cleaned[issueID], blockedBy)
			}
		}
	}

	return cleaned
}

func (c *Model) issueDependencies(projectIssues []redmine.Issue) (map[int][]int, error) {
	dependencies := make(map[int][]int)
	for _, issue := range projectIssues {
		dependencies[issue.Id] = make([]int, 0)
		relations, err := c.Api().IssueRelations(issue.Id)
		if err != nil && err.Error() != "Not Found" {
			return dependencies, err
		}
		fmt.Printf("Issue (%d) Relations: %d\n", issue.Id, len(relations))
		for _, relation := range relations {
			if relation.RelationType != RelationBlocks {
				continue
			}
			fmt.Printf("Issue (%d) - %d is blocked by %d <- needs to be done first\n", issue.Id, relation.IssueToId, relation.IssueId)
			dependencies[relation.IssueToId] = append(dependencies[relation.IssueId], relation.IssueId)
		}
	}

	return dependencies, nil
}

func (c *Model) getCorrectIssue(issues []redmine.Issue, priorities models.Priorities, states models.States) []redmine.Issue {
	valid := make([]redmine.Issue, 0)
	for _, priority := range priorities {
		fmt.Printf("PRIORITY: %q @ %q\n", priority.Type, priority.State)

		state := states.Get(priority.State)
		if !state.AI {
			fmt.Printf("SKIP %q @ %q - NOT FOR AI\n", priority.Type, priority.State)
			continue
		}

		for _, issue := range issues {
			fmt.Printf("ISSUE: %q (%d) - %q\n", issue.Tracker.Name, issue.Id, issue.Status.Name)
			if issue.Tracker.Name != priority.Type {
				fmt.Printf("SKIP %q (%d) - not %q\n", issue.Tracker.Name, issue.Id, priority.Type)
				continue
			}
			if issue.Status.Name != string(priority.State) {
				fmt.Printf("SKIP %q (%d) - %q != %q\n", issue.Tracker.Name, issue.Id, issue.Status.Name, priority.State)
				continue
			}
			fmt.Println("ISSUE:", issue.Tracker.Name, issue.Id)
			valid = append(valid, issue)
		}
	}

	return valid
}
