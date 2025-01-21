package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	queryGetRole    = "SELECT id FROM roles WHERE builtin = 0 AND name = ?"
	queryInsertRole = "INSERT INTO roles (name, position, assignable, builtin, permissions, issues_visibility, users_visibility, time_entries_visibility, all_roles_managed, settings) VALUES (?, 1, 1, 0, ?, 'all', 'all', 'all', 1, ?)"
)

// RoleWorker is a constant for Worker role
const RoleWorker = "Worker"

func (c *Model) DBGetWorkerRole() (int, error) {
	var ids []int
	err := c.queryAndScan(queryGetRole, func(rows *sql.Rows) error {
		var row int
		if err := rows.Scan(&row); err != nil {
			return err
		}
		ids = append(ids, row)
		return nil
	}, RoleWorker)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, err
	}

	for _, id := range ids {
		return id, nil
	}
	return 0, nil
}

func (c *Model) DBCreateWorkerRole() error {
	exists, err := c.checkIfExists(queryGetRole, RoleWorker)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}
	if exists {
		log.Println("Worker Role already exists")
		return nil
	}

	permissionsRaw, settingsRaw := c.getAllPermissionsAndSettings()
	result, err := c.execDML(queryInsertRole, RoleWorker, permissionsRaw, settingsRaw)
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

func (c *Model) getAllPermissionsAndSettings() (permissionsRaw, settingsRaw string) {
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

	permissionsRaw = strings.Join(permissions, "\n")
	settingsRaw = strings.Join(settings, "\n")

	return
}
