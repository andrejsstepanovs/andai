package employee

import (
	"errors"

	"github.com/andrejsstepanovs/andai/internal/exec"
	model "github.com/andrejsstepanovs/andai/internal/redmine"
	redminemodels "github.com/andrejsstepanovs/andai/internal/redmine/models"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/mattn/go-redmine"
)

var ErrNegativeOutcome = errors.New("negative outcome")

type Routine struct {
	model             *model.Model
	llmPool           *settings.LlmModels
	issue             redmine.Issue
	parent            *redmine.Issue
	parents           []redmine.Issue
	closedChildrenIDs []int
	children          []redmine.Issue
	siblings          []redmine.Issue
	project           redmine.Project
	projectCfg        settings.Project
	projectRepo       redminemodels.Repository
	aiderConfig       settings.Aider
	workbench         *exec.Workbench
	state             settings.State
	issueType         settings.IssueType
	issueTypes        settings.IssueTypes
	job               settings.Job
	history           []string
	contextFiles      []string
}

// NewRoutine creates an Routine instance configured to work on a specific Redmine issue.
// It initializes the employee with all necessary context including issue relationships,
// project details, and workflow configuration.
func NewRoutine(
	model *model.Model,
	llmPool *settings.LlmModels,
	issue redmine.Issue,
	parentIssue *redmine.Issue,
	parentIssues []redmine.Issue,
	closedChildrenIDs []int,
	childIssues []redmine.Issue,
	siblingIssues []redmine.Issue,
	project redmine.Project,
	projectConfig settings.Project,
	workbench *exec.Workbench,
	aiderConfig settings.Aider,
	state settings.State,
	issueType settings.IssueType,
	issueTypes settings.IssueTypes,
	projectRepo redminemodels.Repository,
) *Routine {
	return &Routine{
		model:             model,
		llmPool:           llmPool,
		issue:             issue,
		parent:            parentIssue,
		parents:           parentIssues,
		closedChildrenIDs: closedChildrenIDs,
		children:          childIssues,
		siblings:          siblingIssues,
		project:           project,
		projectCfg:        projectConfig,
		workbench:         workbench,
		aiderConfig:       aiderConfig,
		state:             state,
		issueType:         issueType,
		issueTypes:        issueTypes,
		job:               issueType.Jobs.Get(settings.StateName(issue.Status.Name)),
		projectRepo:       projectRepo,
	}
}
