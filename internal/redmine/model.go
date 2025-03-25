package redmine

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

type database interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
}

type api interface {
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
	db  database
	api api
}

func NewModel(db database, api api) *Model {
	return &Model{
		api: api,
		db:  db,
	}
}

func (c *Model) API() api {
	return c.api
}

func (c *Model) DB() database {
	return c.db
}
