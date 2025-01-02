package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	workflow "github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

const ROLE_WORKER = "Worker"
const ADMIN_LOGIN = "admin"

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

func (c *Model) DbGetAllUsers() ([]redmine.User, error) {
	rows, err := c.db.Query("SELECT id, login, firstname, lastname FROM users")
	if err != nil {
		fmt.Println("Redmine Database Ping Fail")
		return nil, fmt.Errorf("error database: %v", err)
	}
	defer rows.Close()

	var users []redmine.User

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var user redmine.User
		if err := rows.Scan(
			&user.Id,
			&user.Login,
			&user.Firstname,
			&user.Lastname,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Model) DbProjectTrackers(projectID int) ([]int, error) {
	rows, err := c.db.Query("SELECT tracker_id FROM projects_trackers WHERE project_id = ?", projectID)
	if err != nil {
		return []int{}, fmt.Errorf("error database: %v", err)
	}
	defer rows.Close()

	var trackerIDs []int
	for rows.Next() {
		var trackerID int
		if err := rows.Scan(&trackerID); err != nil {
			return []int{}, err
		}
		trackerIDs = append(trackerIDs, trackerID)
	}
	if err = rows.Err(); err != nil {
		return []int{}, err
	}

	return trackerIDs, nil
}

func (c *Model) ApiGetUsers() ([]redmine.User, error) {
	users, err := c.api.Users()
	if err != nil {
		return nil, fmt.Errorf("error redmine API get users: %v", err)
	}
	return users, nil
}

func (c *Model) DbUpdateApiToken(userId int, tokenValue string) error {
	result, err := c.db.Exec("UPDATE tokens SET value = ?, updated_on = NOW() WHERE action = ? AND user_id = ?", tokenValue, "api", userId)
	if err != nil {
		return fmt.Errorf("update settings token db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("token not updated")
	}
	return nil
}

func (c *Model) DbCreateApiToken(userId int, tokenValue string) error {
	result, err := c.db.Exec("INSERT INTO tokens (value, action, user_id, created_on, updated_on) VALUES (?, ?, ?, NOW(), NOW())", tokenValue, "api", userId)
	if err != nil {
		return fmt.Errorf("insert settings token db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("token not created")
	}
	return nil
}

func (c *Model) DbGetToken(userId int) (models.Token, error) {
	results, err := c.db.Query("SELECT id, action, value FROM tokens WHERE action = ? AND user_id = ?", "api", userId)
	if err != nil {
		return models.Token{}, fmt.Errorf("error redmine admin: %v", err)
	}
	defer results.Close()

	var rows []models.Token
	for results.Next() {
		var row models.Token
		if err := results.Scan(&row.ID, &row.Action, &row.Value); err != nil {
			return models.Token{}, err
		}
		rows = append(rows, row)
	}
	err = results.Err()
	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return models.Token{}, err
	}
	for _, row := range rows {
		return row, nil
	}
	return models.Token{}, nil
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

func (c *Model) DbSettingsAdminMustNotChangePassword() error {
	result, err := c.db.Exec("UPDATE users SET must_change_passwd = 0 WHERE login = ?", "admin")
	if err != nil {
		return fmt.Errorf("update users db err: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}

	if affected > 0 {
		fmt.Println("Admin must_change_passwd set to 0")
	} else {
		fmt.Println("Admin must_change_passwd not changed")
	}

	return nil
}

func (c *Model) DbGetSettings(settingName string) ([]models.Settings, error) {
	results, err := c.db.Query("SELECT id, name, value FROM settings WHERE name = ?", settingName)
	if err != nil {
		return nil, fmt.Errorf("db settings err: %v", err)
	}
	defer results.Close()

	var rows []models.Settings
	for results.Next() {
		var row models.Settings
		if err := results.Scan(&row.ID, &row.Name, &row.Value); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	err = results.Err()
	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return nil, err
	}
	return rows, nil
}

func (c *Model) DbSettingsInsert(settingName, value string) error {
	result, err := c.db.Exec("INSERT INTO settings (name, value, updated_on) VALUES (?, ?, NOW())", settingName, value)
	if err != nil {
		return fmt.Errorf("update settings db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("Admin rest_api_enabled not changed")
	}
	return nil
}

func (c *Model) DbSettingsUpdate(settingName, value string) error {
	result, err := c.db.Exec("UPDATE settings SET value = ?, updated_on = NOW() WHERE name = ?", value, settingName)
	if err != nil {
		return fmt.Errorf("update settings db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected > 0 {
		return nil
	}
	return nil
}

func (c *Model) DbSettingsEnableAPI() error {
	const SETTING_NAME = "rest_api_enabled"
	const SETTING_VALUE = "1"

	rows, err := c.DbGetSettings(SETTING_NAME)
	for _, row := range rows {
		fmt.Printf("Setting ID: %d, Name: %s, Value: %s\n", row.ID, row.Name, row.Value)
		if row.Value != SETTING_VALUE {
			fmt.Printf("Setting %s is not enabled\n", SETTING_NAME)
			err = c.DbSettingsUpdate(SETTING_NAME, SETTING_VALUE)
			if err != nil {
				return fmt.Errorf("update settings db err: %v", err)
			}
			fmt.Printf("Setting %s updated to %s\n", SETTING_NAME, SETTING_VALUE)
			return nil
		}
		fmt.Printf("Setting %s is OK\n", SETTING_NAME)
		return nil
	}

	fmt.Printf("Setting %s is not present\n", SETTING_NAME)
	err = c.DbSettingsInsert(SETTING_NAME, SETTING_VALUE)
	if err != nil {
		return fmt.Errorf("insert settings db err: %v", err)
	}
	fmt.Printf("Setting %s created with value: %s\n", SETTING_NAME, SETTING_VALUE)
	return nil
}

func (c *Model) DbSaveWiki(project redmine.Project, content string) error {
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

func (c *Model) GetProjects() ([]redmine.Project, error) {
	projects, err := c.api.Projects()
	if err != nil {
		return nil, fmt.Errorf("error redmine projects: %v", err)
	}
	return projects, nil
}

func (c *Model) SaveProject(project redmine.Project) (error, redmine.Project) {
	projects, err := c.api.Projects()
	if err != nil {
		return fmt.Errorf("error redmine projects: %v", err), redmine.Project{}
	}

	for _, p := range projects {
		fmt.Println(fmt.Sprintf("ID: %d, Name: %s", p.Id, p.Name))
		if p.Identifier == viper.GetString("project.id") {
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

func (c *Model) DbGetRepository(project redmine.Project) (models.Repository, error) {
	results, err := c.db.Query("SELECT id, project_id, root_url FROM repositories WHERE identifier = ?", project.Identifier)
	if err != nil {
		return models.Repository{}, fmt.Errorf("error redmine admin: %v", err)
	}
	defer results.Close()

	var rows []models.Repository
	for results.Next() {
		var row models.Repository
		if err := results.Scan(&row.ID, &row.ProjectID, &row.RootUrl); err != nil {
			return models.Repository{}, err
		}
		rows = append(rows, row)
	}
	err = results.Err()
	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return models.Repository{}, err
	}
	for _, row := range rows {
		return row, nil
	}
	return models.Repository{}, nil
}

func (c *Model) DbSaveProjectTrackers(project redmine.Project) error {
	allTrackers, err := c.api.Trackers()
	if err != nil {
		return fmt.Errorf("error redmine trackers: %v", err)
	}

	existingTrackerIDs, err := c.DbProjectTrackers(project.Id)
	if err != nil {
		return fmt.Errorf("get project trackers for project %d err: %v", project.Id, err)
	}

	createTrackers := make([]redmine.IdName, 0)
	for _, tracker := range allTrackers {
		exists := false
		for _, existingTrackerID := range existingTrackerIDs {
			if tracker.Id == existingTrackerID {
				fmt.Printf("Project %q Tracker for %q already exists ID: %d\n", project.Identifier, tracker.Name, tracker.Id)
				exists = true
				break
			}
		}
		if !exists {
			createTrackers = append(createTrackers, tracker)
		}
	}

	for _, tracker := range createTrackers {
		fmt.Println(fmt.Sprintf("Tracker: %s", tracker.Name))
		err = c.InsertProjectTracker(project.Id, tracker.Id)
		if err != nil {
			return fmt.Errorf("redmine project tracker insert err: %v", err)
		}
	}
	return nil
}

func (c *Model) DbSaveGit(project redmine.Project, gitPath string) error {
	newUrl := fmt.Sprintf("%s/%s", strings.TrimRight(viper.GetString("redmine.repositories"), "/"), strings.TrimLeft(gitPath, "/"))

	repository, err := c.DbGetRepository(project)
	if err != nil {
		return fmt.Errorf("redmine get repository err: %v", err)
	}
	if repository.ID > 0 {
		fmt.Printf("ID: %d, ProjectID: %s, Url: %s\n", repository.ID, repository.ProjectID, repository.RootUrl)
		if repository.RootUrl == newUrl {
			return nil
		}
		fmt.Println("Mismatch Repository root_url")
		result, err := c.db.Exec("UPDATE repositories SET root_url = ?, created_on = NOW() WHERE id = ?", newUrl, repository.ID)
		if err != nil {
			return fmt.Errorf("update repository db err: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("rows affected err: %v", err)
		}
		if affected == 0 {
			return errors.New("project repository root_url not changed")
		}
		fmt.Println("Project repository root_url updated")
		return nil
	}

	query := "INSERT INTO repositories (project_id, root_url, type, path_encoding, extra_info, identifier, is_default, created_on) VALUES(?, ?, 'Repository::Git', 'UTF-8', ?, ?, 1, NOW())"

	result, err := c.db.Exec(query, project.Id, newUrl, "---\nextra_report_last_commit: '0'\n", project.Identifier)
	if err != nil {
		return fmt.Errorf("error redmine git save: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("project repository not saved")
	}

	fmt.Println("project repository inserted")
	return nil
}

func (c *Model) InsertProjectTracker(projectID, trackerID int) error {
	query := "INSERT INTO projects_trackers (project_id, tracker_id) VALUES(?, ?)"

	result, err := c.db.Exec(query, projectID, trackerID)
	if err != nil {
		return fmt.Errorf("error redmine project tracker insert: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("project tracker not saved")
	}

	fmt.Println("project tracker inserted")
	return nil
}

func (c *Model) SaveTrackers(trackers workflow.IssueTypes, defaultStatus redmine.IssueStatus) error {
	current, err := c.api.Trackers()
	if err != nil {
		return fmt.Errorf("error redmine trackers: %v", err)
	}

	newTrackers := make([]workflow.IssueType, 0)
	for _, t := range trackers {
		exists := false
		for _, ct := range current {
			if ct.Name == string(t.Name) {
				fmt.Printf("Tracker %s already exists: %d\n", ct.Name, ct.Id)
				exists = true
				break
			}
		}
		if !exists {
			fmt.Println(fmt.Sprintf("Tracker: %s", t.Name))
			newTrackers = append(newTrackers, t)
		}
	}

	if len(newTrackers) == 0 {
		fmt.Println("Trackers OK")
		return nil
	}

	for i, tracker := range newTrackers {
		fmt.Println(fmt.Sprintf("Creating New Tracker: %s", tracker.Name))
		err = c.InsertTracker(tracker, i+1, defaultStatus.Id)
		if err != nil {
			return fmt.Errorf("redmine tracker insert err: %v", err)
		}
	}

	return nil
}

func (c *Model) APIGetDefaultStatus() (redmine.IssueStatus, error) {
	statuses, err := c.api.IssueStatuses()
	if err != nil {
		return redmine.IssueStatus{}, fmt.Errorf("error redmine issue status: %v", err)
	}

	for _, status := range statuses {
		return status, nil
	}

	return redmine.IssueStatus{}, errors.New("default status not found")
}

func (c *Model) InsertTracker(issueType workflow.IssueType, position, defaultState int) error {
	query := "INSERT INTO trackers (name, description, position, default_status_id) VALUES (?, ?, ?, ?)"
	result, err := c.db.Exec(query, issueType.Name, issueType.Description, position, defaultState)
	if err != nil {
		return fmt.Errorf("redmine tracker err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("tracker not created")
	}
	return nil
}

func (c *Model) SaveIssueStatuses(statuses []redmine.IssueStatus) error {
	current, err := c.api.IssueStatuses()
	if err != nil {
		return fmt.Errorf("error redmine issue status: %v", err)
	}

	newStatuses := make([]redmine.IssueStatus, 0)
	for _, status := range statuses {
		exists := false
		for _, s := range current {
			if s.Name == status.Name {
				exists = true
				//fmt.Printf("Issue status %s already exists: %s\n", s.Name, s.Id)
				break
			}
		}
		if !exists {
			newStatuses = append(newStatuses, status)
		}
	}

	if len(newStatuses) == 0 {
		fmt.Println("Issue Statuses OK")
		return nil
	}

	for i, status := range newStatuses {
		fmt.Println(fmt.Sprintf("Creating New Issue Status: %s", status.Name))
		err = c.InsertIssueStatus(status, i+1)
		if err != nil {
			return fmt.Errorf("redmine issue status insert err: %v", err)
		}
	}

	return nil
}

func (c *Model) InsertIssueStatus(status redmine.IssueStatus, position int) error {
	result, err := c.db.Exec("INSERT INTO issue_statuses (name, is_closed, position) VALUES (?, ?, ?)", status.Name, status.IsClosed, position)
	if err != nil {
		return fmt.Errorf("error redmine issue status insert: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("token not created")
	}
	return nil
}

func (c *Model) DBGetWorkerRole() (int, error) {
	results, err := c.db.Query("SELECT id FROM roles WHERE builtin = 0 AND name = ?", ROLE_WORKER)
	if err != nil {
		return 0, fmt.Errorf("redmine db err: %v", err)
	}
	defer results.Close()

	var rows []int
	for results.Next() {
		var row int
		if err := results.Scan(&row); err != nil {
			return 0, err
		}
		rows = append(rows, row)
	}
	err = results.Err()
	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, err
	}
	for _, row := range rows {
		return row, nil
	}
	return 0, nil
}

func (c *Model) DBCreateWorkerRole() error {
	permissions := []string{
		"---",
		"- :add_project",
		"- :edit_project",
		"- :close_project",
		"- :delete_project",
		"- :select_project_publicity",
		"- :select_project_modules",
		"- :manage_members",
		"- :manage_versions",
		"- :add_subprojects",
		"- :manage_public_queries",
		"- :save_queries",
		"- :view_messages",
		"- :add_messages",
		"- :edit_messages",
		"- :edit_own_messages",
		"- :delete_messages",
		"- :delete_own_messages",
		"- :view_message_watchers",
		"- :add_message_watchers",
		"- :delete_message_watchers",
		"- :manage_boards",
		"- :view_calendar",
		"- :view_documents",
		"- :add_documents",
		"- :edit_documents",
		"- :delete_documents",
		"- :view_files",
		"- :manage_files",
		"- :view_gantt",
		"- :view_issues",
		"- :add_issues",
		"- :edit_issues",
		"- :edit_own_issues",
		"- :copy_issues",
		"- :manage_issue_relations",
		"- :manage_subtasks",
		"- :set_issues_private",
		"- :set_own_issues_private",
		"- :add_issue_notes",
		"- :edit_issue_notes",
		"- :edit_own_issue_notes",
		"- :view_private_notes",
		"- :set_notes_private",
		"- :delete_issues",
		"- :view_issue_watchers",
		"- :add_issue_watchers",
		"- :delete_issue_watchers",
		"- :import_issues",
		"- :manage_categories",
		"- :view_news",
		"- :manage_news",
		"- :comment_news",
		"- :view_changesets",
		"- :browse_repository",
		"- :commit_access",
		"- :manage_related_issues",
		"- :manage_repository",
		"- :view_time_entries",
		"- :log_time",
		"- :edit_time_entries",
		"- :edit_own_time_entries",
		"- :manage_project_activities",
		"- :log_time_for_other_users",
		"- :import_time_entries",
		"- :view_wiki_pages",
		"- :view_wiki_edits",
		"- :export_wiki_pages",
		"- :edit_wiki_pages",
		"- :rename_wiki_pages",
		"- :delete_wiki_pages",
		"- :delete_wiki_pages_attachments",
		"- :view_wiki_page_watchers",
		"- :add_wiki_page_watchers",
		"- :delete_wiki_page_watchers",
		"- :protect_wiki_pages",
		"- :manage_wiki",
		"",
	}
	settings := []string{
		"---",
		"permissions_all_trackers:",
		"  view_issues: '1'",
		"  add_issues: '1'",
		"  edit_issues: '1'",
		"  add_issue_notes: '1'",
		"  delete_issues: '1'",
		"permissions_tracker_ids:",
		"  view_issues: []",
		"  add_issues: []",
		"  edit_issues: []",
		"  add_issue_notes: []",
		"  delete_issues: []",
		"",
	}
	permissionsRaw := strings.Join(permissions, "\n")
	settingsRaw := strings.Join(settings, "\n")

	query := "INSERT INTO roles (name, position, assignable, builtin, permissions, issues_visibility, users_visibility, time_entries_visibility, all_roles_managed, settings) VALUES (?, 1, 1, 0, ?, 'all', 'all', 'all', 1, ?);"

	result, err := c.db.Exec(query, ROLE_WORKER, permissionsRaw, settingsRaw)
	if err != nil {
		return fmt.Errorf("insert role err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("roles rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("role not updated")
	}
	return nil
}

func (c *Model) DBInsertWorkflow(trackerID, oldStatusID, newStatusID, roleID int) error {
	query := "INSERT INTO workflows (tracker_id, old_status_id, new_status_id, role_id, type) VALUES (?, ?, ?, ?, 'WorkflowTransition')"
	result, err := c.db.Exec(query, trackerID, oldStatusID, newStatusID, roleID)
	if err != nil {
		return fmt.Errorf("insert workflow err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("workflow row affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("workflow not updated")
	}
	return nil
}

func (c *Model) DBGetWorkflows(trackerID int) ([]models.Workflow, error) {
	rows, err := c.db.Query("SELECT id, tracker_id, old_status_id, new_status_id, role_id, assignee, author, type, field_name, rule FROM workflows WHERE tracker_id = ?", trackerID)
	if err != nil {
		return nil, fmt.Errorf("error database: %v", err)
	}
	defer rows.Close()

	var workflows []models.Workflow
	for rows.Next() {
		var row models.Workflow
		if err := rows.Scan(
			&row.ID,
			&row.TrackerID,
			&row.OldStatusID,
			&row.NewStatusID,
			&row.RoleID,
			&row.Assignee,
			&row.Author,
			&row.Type,
			&row.FieldName,
			&row.Rule,
		); err != nil {
			return nil, err
		}
		workflows = append(workflows, row)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workflows, nil
}
