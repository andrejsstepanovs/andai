package redmine

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
)

type database interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
}

type api interface {
	Users() ([]redmine.User, error)
	WikiPage(projectId int, title string) (*redmine.WikiPage, error)
	CreateWikiPage(projectId int, wikiPage redmine.WikiPage) (*redmine.WikiPage, error)
	UpdateWikiPage(projectId int, wikiPage redmine.WikiPage) error
	Projects() ([]redmine.Project, error)
	UpdateProject(project redmine.Project) error
	CreateProject(project redmine.Project) (*redmine.Project, error)
	IssueStatuses() ([]redmine.IssueStatus, error)
	Trackers() ([]redmine.IdName, error)
	IssueRelations(issueId int) ([]redmine.IssueRelation, error)
	IssuesOf(projectId int) ([]redmine.Issue, error)
	Issue(id int) (*redmine.Issue, error)
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

func (c *Model) Api() api {
	return c.api
}

func (c *Model) Db() database {
	return c.db
}
