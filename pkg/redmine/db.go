package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	workflow "github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

const (
	queryGetAllUsers             = "SELECT id, login, firstname, lastname FROM users"
	queryGetProjectTrackers      = "SELECT tracker_id FROM projects_trackers WHERE project_id = ?"
	queryUpdateTokens            = "UPDATE tokens SET value = ?, updated_on = NOW() WHERE action = ? AND user_id = ?"
	queryInsertTokens            = "INSERT INTO tokens (value, action, user_id, created_on, updated_on) VALUES (?, ?, ?, NOW(), NOW())"
	queryGetToken                = "SELECT id, action, value FROM tokens WHERE action = ? AND user_id = ?"
	queryUpdateAdminNoChangePass = "UPDATE users SET must_change_passwd = 0 WHERE login = ?"
	queryGetSettings             = "SELECT id, name, value FROM settings WHERE name = ?"
	queryInsertSettings          = "INSERT INTO settings (name, value, updated_on) VALUES (?, ?, NOW())"
	queryUpdateSettingsValue     = "UPDATE settings SET value = ?, updated_on = NOW() WHERE name = ?"
	queryGetProjectRepository    = "SELECT id, project_id, root_url FROM repositories WHERE identifier = ?"
	queryInsertRepository        = "INSERT INTO repositories (project_id, root_url, type, path_encoding, extra_info, identifier, is_default, created_on) VALUES(?, ?, 'Repository::Git', 'UTF-8', ?, ?, 1, NOW())"
	queryUpdateRepository        = "UPDATE repositories SET root_url = ?, created_on = NOW() WHERE id = ?"
	queryInsertProjectTracker    = "INSERT INTO projects_trackers (project_id, tracker_id) VALUES(?, ?)"
	queryInsertTracker           = "INSERT INTO trackers (name, description, position, default_status_id) VALUES (?, ?, ?, ?)"
	queryGetIssuePriority        = "SELECT id FROM enumerations WHERE type = 'IssuePriority' AND is_default = 1 AND name = ?"
	queryInsertIssuePriority     = "INSERT INTO enumerations (name, position, is_default, type, active, project_id, parent_id, position_name) VALUES (?, 1, 1, 'IssuePriority', 1, NULL, NULL, 'default')"
	queryInsertIssueStatus       = "INSERT INTO issue_statuses (name, is_closed, position) VALUES (?, ?, ?)"
	queryGetRole                 = "SELECT id FROM roles WHERE builtin = 0 AND name = ?"
	queryInsertRole              = "INSERT INTO roles (name, position, assignable, builtin, permissions, issues_visibility, users_visibility, time_entries_visibility, all_roles_managed, settings) VALUES (?, 1, 1, 0, ?, 'all', 'all', 'all', 1, ?)"
	queryInsertWorkflow          = "INSERT INTO workflows (tracker_id, old_status_id, new_status_id, role_id, type) VALUES (?, ?, ?, ?, 'WorkflowTransition')"
	queryGetWorkflow             = "SELECT id, tracker_id, old_status_id, new_status_id, role_id, assignee, author, type, field_name, rule FROM workflows WHERE tracker_id = ?"
)

func (c *Model) DbGetAllUsers() ([]redmine.User, error) {
	var users []redmine.User
	err := c.queryAndScan(queryGetAllUsers, func(rows *sql.Rows) error {
		var user redmine.User
		if err := rows.Scan(&user.Id, &user.Login, &user.Firstname, &user.Lastname); err != nil {
			return err
		}
		users = append(users, user)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Model) DbProjectTrackers(projectID int) ([]int, error) {
	var trackerIDs []int
	err := c.queryAndScan(queryGetProjectTrackers, func(rows *sql.Rows) error {
		var trackerID int
		if err := rows.Scan(&trackerID); err != nil {
			return err
		}
		trackerIDs = append(trackerIDs, trackerID)
		return nil
	}, projectID)

	if err != nil {
		return nil, err
	}

	return trackerIDs, nil
}

func (c *Model) DbUpdateApiToken(userId int, tokenValue string) error {
	result, err := c.execDML(queryUpdateTokens, tokenValue, "api", userId)
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
	result, err := c.execDML(queryInsertTokens, tokenValue, "api", userId)
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
	var tokens []models.Token
	err := c.queryAndScan(queryGetToken, func(rows *sql.Rows) error {
		var token models.Token
		if err := rows.Scan(&token.ID, &token.Action, &token.Value); err != nil {
			return err
		}
		tokens = append(tokens, token)
		return nil
	}, "api", userId)

	if err != nil {
		return models.Token{}, err
	}

	if len(tokens) > 0 {
		return tokens[0], nil
	}
	return models.Token{}, nil
}

func (c *Model) DbSettingsAdminMustNotChangePassword() error {
	result, err := c.execDML(queryUpdateAdminNoChangePass, "admin")
	if err != nil {
		return fmt.Errorf("update users db err: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}

	if affected > 0 {
		log.Println("Admin must_change_passwd set to 0")
	} else {
		log.Println("Admin must_change_passwd not changed")
	}

	return nil
}

func (c *Model) DbGetSettings(settingName string) ([]models.Settings, error) {
	var settings []models.Settings
	err := c.queryAndScan(queryGetSettings, func(rows *sql.Rows) error {
		var row models.Settings
		if err := rows.Scan(&row.ID, &row.Name, &row.Value); err != nil {
			return err
		}
		settings = append(settings, row)
		return nil
	}, settingName)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return nil, err
	}
	return settings, nil
}

func (c *Model) DbSettingsInsert(settingName, value string) error {
	result, err := c.execDML(queryInsertSettings, settingName, value)
	if err != nil {
		return fmt.Errorf("update settings db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("admin rest_api_enabled not changed")
	}
	return nil
}

func (c *Model) DbSettingsUpdate(settingName, value string) error {
	result, err := c.execDML(queryUpdateSettingsValue, value, settingName)
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
	const settingsName = "rest_api_enabled"
	const settingsValue = "1"

	rows, err := c.DbGetSettings(settingsName)
	for _, row := range rows {
		log.Printf("Setting Identifier: %d, Name: %s, Value: %s\n", row.ID, row.Name, row.Value)
		if row.Value != settingsValue {
			log.Printf("Setting %s is not enabled\n", settingsName)
			err = c.DbSettingsUpdate(settingsName, settingsValue)
			if err != nil {
				return fmt.Errorf("update settings db err: %v", err)
			}
			log.Printf("Setting %s updated to %s\n", settingsName, settingsValue)
			return nil
		}
		log.Printf("Setting %s is OK\n", settingsName)
		return nil
	}

	log.Printf("Setting %s is not present\n", settingsName)
	err = c.DbSettingsInsert(settingsName, settingsValue)
	if err != nil {
		return fmt.Errorf("insert settings db err: %v", err)
	}
	log.Printf("Setting %s created with value: %s\n", settingsName, settingsValue)
	return nil
}

func (c *Model) DBSaveTrackers(trackers workflow.IssueTypes, defaultStatus redmine.IssueStatus, current []redmine.IdName) error {
	newTrackers := make([]workflow.IssueType, 0)
	for _, t := range trackers {
		exists := false
		for _, ct := range current {
			if ct.Name == string(t.Name) {
				log.Printf("Tracker %s already exists: %d\n", ct.Name, ct.Id)
				exists = true
				break
			}
		}
		if !exists {
			log.Println(fmt.Sprintf("Tracker: %s", t.Name))
			newTrackers = append(newTrackers, t)
		}
	}

	if len(newTrackers) == 0 {
		log.Println("Trackers OK")
		return nil
	}

	for i, tracker := range newTrackers {
		log.Println(fmt.Sprintf("Creating New Tracker: %s", tracker.Name))
		err := c.DBInsertTracker(tracker, i+1, defaultStatus.Id)
		if err != nil {
			return fmt.Errorf("redmine tracker insert err: %v", err)
		}
	}

	return nil
}

func (c *Model) DbGetRepository(project redmine.Project) (models.Repository, error) {
	var repos []models.Repository
	err := c.queryAndScan(queryGetProjectRepository, func(rows *sql.Rows) error {
		var row models.Repository
		if err := rows.Scan(&row.ID, &row.ProjectID, &row.RootUrl); err != nil {
			return err
		}
		repos = append(repos, row)
		return nil
	}, project.Identifier)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return models.Repository{}, err
	}
	for _, row := range repos {
		return row, nil
	}
	return models.Repository{}, nil
}

func (c *Model) DbSaveProjectTrackers(project redmine.Project, allTrackers []redmine.IdName) error {
	existingTrackerIDs, err := c.DbProjectTrackers(project.Id)
	if err != nil {
		return fmt.Errorf("get project trackers for project %d err: %v", project.Id, err)
	}

	createTrackers := make([]redmine.IdName, 0)
	for _, tracker := range allTrackers {
		exists := false
		for _, existingTrackerID := range existingTrackerIDs {
			if tracker.Id == existingTrackerID {
				log.Printf("Project %q Tracker for %q already exists Identifier: %d\n", project.Identifier, tracker.Name, tracker.Id)
				exists = true
				break
			}
		}
		if !exists {
			createTrackers = append(createTrackers, tracker)
		}
	}

	for _, tracker := range createTrackers {
		log.Println(fmt.Sprintf("Tracker: %s", tracker.Name))
		err = c.DBInsertProjectTracker(project.Id, tracker.Id)
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
		log.Printf("Repository ID: %d, ProjectID: %s, Url: %s\n", repository.ID, repository.ProjectID, repository.RootUrl)
		if repository.RootUrl == newUrl {
			return nil
		}
		log.Println("Mismatch Repository root_url")
		result, err := c.execDML(queryUpdateRepository, newUrl, repository.ID)
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
		log.Printf("Project %s repository root_url updated\n", project.Identifier)
		return nil
	}

	result, err := c.execDML(queryInsertRepository, project.Id, newUrl, "---\nextra_report_last_commit: '0'\n", project.Identifier)
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

	log.Println("project repository inserted")
	return nil
}

func (c *Model) DBInsertProjectTracker(projectID, trackerID int) error {
	result, err := c.execDML(queryInsertProjectTracker, projectID, trackerID)
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

	log.Println("project tracker inserted")
	return nil
}

func (c *Model) DBInsertTracker(issueType workflow.IssueType, position, defaultState int) error {
	result, err := c.execDML(queryInsertTracker, issueType.Name, issueType.Description, position, defaultState)
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

func (c *Model) GetDefaultNormalPriority() (int, error) {
	var ids []int
	err := c.queryAndScan(queryGetIssuePriority, func(rows *sql.Rows) error {
		var row int
		if err := rows.Scan(&row); err != nil {
			return err
		}
		ids = append(ids, row)
		return nil
	}, ISSUE_PRIORITY)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, err
	}
	for _, id := range ids {
		return id, nil
	}
	return 0, nil
}

func (c *Model) DBInsertDefaultNormalPriority() error {
	result, err := c.execDML(queryInsertIssuePriority, ISSUE_PRIORITY)
	if err != nil {
		return fmt.Errorf("redmine issue priority err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("issue priority not created")
	}
	return nil
}

func (c *Model) DBSaveIssueStatuses(statuses []redmine.IssueStatus, current []redmine.IssueStatus) error {
	newStatuses := make([]redmine.IssueStatus, 0)
	for _, status := range statuses {
		exists := false
		for _, s := range current {
			if s.Name == status.Name {
				exists = true
				break
			}
		}
		if !exists {
			newStatuses = append(newStatuses, status)
		}
	}

	if len(newStatuses) == 0 {
		log.Println("Issue Statuses OK")
		return nil
	}

	for i, status := range newStatuses {
		log.Println(fmt.Sprintf("Creating New Issue Status: %s", status.Name))
		err := c.DBInsertIssueStatus(status, i+1)
		if err != nil {
			return fmt.Errorf("redmine issue status insert err: %v", err)
		}
	}

	return nil
}

func (c *Model) DBInsertIssueStatus(status redmine.IssueStatus, position int) error {
	result, err := c.execDML(queryInsertIssueStatus, status.Name, status.IsClosed, position)
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
	var ids []int
	err := c.queryAndScan(queryGetRole, func(rows *sql.Rows) error {
		var row int
		if err := rows.Scan(&row); err != nil {
			return err
		}
		ids = append(ids, row)
		return nil
	}, ROLE_WORKER)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, err
	}

	for _, id := range ids {
		return id, nil
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

	exists, err := c.checkIfExists(queryGetRole, ROLE_WORKER)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}
	if exists {
		log.Println("Worker Role already exists")
		return nil
	}

	result, err := c.execDML(queryInsertRole, ROLE_WORKER, permissionsRaw, settingsRaw)
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
	result, err := c.execDML(queryInsertWorkflow, trackerID, oldStatusID, newStatusID, roleID)
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
	var workflows []models.Workflow
	err := c.queryAndScan(queryGetWorkflow, func(rows *sql.Rows) error {
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
			return err
		}
		workflows = append(workflows, row)
		return nil
	}, trackerID)

	if err != nil {
		return nil, err
	}

	return workflows, nil
}
