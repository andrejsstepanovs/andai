package redmine

import (
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
)

func (c *Model) ApiGetUsers() ([]redmine.User, error) {
	users, err := c.api.Users()
	if err != nil {
		return nil, fmt.Errorf("error redmine API get users: %v", err)
	}
	return users, nil
}

func (c *Model) ApiAdmin() (redmine.User, error) {
	users, err := c.DbGetAllUsers()
	if err != nil {
		return redmine.User{}, fmt.Errorf("error redmine db get users: %v", err)
	}
	for _, user := range users {
		if user.Login == ADMIN_LOGIN {
			return user, nil
		}
	}
	return redmine.User{}, errors.New("admin not found")
}

func (c *Model) GetProjects() ([]redmine.Project, error) {
	projects, err := c.api.Projects()
	if err != nil {
		return nil, fmt.Errorf("error redmine projects: %v", err)
	}
	return projects, nil
}

func (c *Model) SaveProject(project redmine.Project) (error, redmine.Project) {
	current, err := c.api.Projects()
	if err != nil {
		return fmt.Errorf("redmine api projects err: %v", err), redmine.Project{}
	}

	for _, p := range current {
		fmt.Printf("ID: %d, Name: %s Identifier: %s\n", p.Id, p.Name, p.Identifier)
		if p.Identifier == project.Identifier {
			fmt.Printf("Project already exists: %s\n", p.Name)
			project.Id = p.Id
			err = c.api.UpdateProject(project)
			if err != nil && err.Error() != "EOF" {
				fmt.Println("Redmine Update Project Failed")
				return fmt.Errorf("error redmine update project: %v", err.Error()), redmine.Project{}
			}
			fmt.Printf("Project updated: %s\n", project.Name)
			return nil, project
		}
	}

	response, err := c.api.CreateProject(project)
	if err != nil {
		return fmt.Errorf("error redmine create project: '%s'", err.Error()), redmine.Project{}
	}

	return nil, *response
}

func (c *Model) APIGetIssueStatusByName(name string) (redmine.IssueStatus, error) {
	statuses, err := c.api.IssueStatuses()
	if err != nil {
		return redmine.IssueStatus{}, fmt.Errorf("error redmine issue status: %v", err)
	}

	for _, status := range statuses {
		if status.Name == name {
			return status, nil
		}
	}

	return redmine.IssueStatus{}, errors.New("default status not found")
}
