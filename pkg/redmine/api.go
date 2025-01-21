package redmine

import (
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

// AdminLogin is the default admin login
const AdminLogin = "admin"

func (c *Model) APIGetUsers() ([]redmine.User, error) {
	users, err := c.API().Users()
	if err != nil {
		return nil, fmt.Errorf("error redmine API get users: %v", err)
	}
	return users, nil
}

func (c *Model) APIAdmin() (redmine.User, error) {
	users, err := c.DBGetAllUsers()
	if err != nil {
		return redmine.User{}, fmt.Errorf("error redmine db get users: %v", err)
	}
	for _, user := range users {
		if user.Login == AdminLogin {
			return user, nil
		}
	}
	return redmine.User{}, errors.New("admin not found")
}

func (c *Model) APIGetProjects() ([]redmine.Project, error) {
	projects, err := c.API().Projects()
	if err != nil {
		return nil, fmt.Errorf("error redmine projects: %v", err)
	}
	return projects, nil
}

func (c *Model) APISaveWiki(project redmine.Project, content string) error {
	const TITLE = "Wiki"
	content = strings.TrimSpace(content)

	page, err := c.api.WikiPage(project.Id, TITLE)
	if err != nil {
		if err.Error() != "Not Found" {
			return fmt.Errorf("error redmine wiki page: %v", err)
		}

		page = &redmine.WikiPage{Title: TITLE, Text: content}
		_, err = c.api.CreateWikiPage(project.Id, *page)
		if err != nil {
			return fmt.Errorf("error redmine wiki page create: %v", err)
		}
		return nil
	}

	page.Text = content
	err = c.api.UpdateWikiPage(project.Id, *page)
	if err != nil && err.Error() != "EOF" {
		return fmt.Errorf("error redmine wiki page update: %v", err)
	}

	return nil
}

func (c *Model) APISaveProject(project redmine.Project) (redmine.Project, error) {
	current, err := c.API().Projects()
	if err != nil {
		return redmine.Project{}, fmt.Errorf("redmine api projects err: %v", err)
	}

	for _, p := range current {
		log.Printf("ID: %d, Name: %s Identifier: %s\n", p.Id, p.Name, p.Identifier)
		if p.Identifier == project.Identifier {
			log.Printf("Project already exists: %s\n", p.Name)
			project.Id = p.Id
			err = c.API().UpdateProject(project)
			if err != nil && err.Error() != "EOF" {
				log.Println("Redmine Update Project Failed")
				return redmine.Project{}, fmt.Errorf("error redmine update project: %v", err.Error())
			}
			log.Printf("Project updated: %s\n", project.Name)
			return project, nil
		}
	}

	response, err := c.API().CreateProject(project)
	if err != nil {
		return redmine.Project{}, fmt.Errorf("error redmine create project: '%s'", err.Error())
	}

	return *response, nil
}

func (c *Model) APIGetIssueStatusByName(name string) (redmine.IssueStatus, error) {
	statuses, err := c.API().IssueStatuses()
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
