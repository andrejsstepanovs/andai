package redmine

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

func (c *Model) APIGetIssueStatus(name string) (redmine.IssueStatus, error) {
	issuesStatuses, err := c.API().IssueStatuses()
	if err != nil {
		return redmine.IssueStatus{}, fmt.Errorf("error redmine issue status: %v", err)
	}

	for _, issueStatus := range issuesStatuses {
		if issueStatus.Name == name {
			return issueStatus, nil
		}
	}

	return redmine.IssueStatus{}, nil
}
