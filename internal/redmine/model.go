package redmine

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

type DatabaseInterface interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
}

type APIInterface interface {
	Users() ([]redmine.User, error)
	WikiPage(projectID int, title string) (*redmine.WikiPage, error)
	CreateWikiPage(projectID int, wikiPage redmine.WikiPage) (*redmine.WikiPage, error)
	UpdateWikiPage(projectID int, wikiPage redmine.WikiPage) error
	Projects() ([]redmine.Project, error)
	UpdateProject(project redmine.Project) error
	CreateProject(project redmine.Project) (*redmine.Project, error)
	IssueStatuses() ([]redmine.IssueStatus, error)
	Trackers() ([]redmine.IdName, error)
	IssueRelations(issueID int) ([]redmine.IssueRelation, error)
	CreateIssueRelation(issueRelation redmine.IssueRelation) (*redmine.IssueRelation, error)
	IssuesOf(projectID int) ([]redmine.Issue, error)
	Issue(id int) (*redmine.Issue, error)
	Project(id int) (*redmine.Project, error)
	UpdateIssue(issue redmine.Issue) error
	CreateIssue(issue redmine.Issue) (*redmine.Issue, error)
}

type Model struct {
	db  DatabaseInterface
	api APIInterface
}

func NewModel(db DatabaseInterface, api APIInterface) *Model {
	return &Model{
		api: api,
		db:  db,
	}
}

func (c *Model) API() APIInterface {
	return c.api
}

func (c *Model) DB() DatabaseInterface {
	return c.db
}
