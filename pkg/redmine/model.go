package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

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

func (c *Model) DbSaveGit(project redmine.Project, gitPath string) error {
	newUrl := fmt.Sprintf("%s/%s", strings.TrimRight(viper.GetString("redmine.repositories"), "/"), strings.TrimLeft(gitPath, "/"))

	repository, err := c.DbGetRepository(project)
	if err != nil {
		return fmt.Errorf("redmine get repository err: %v", err)
	}
	if repository.ID > 0 {
		fmt.Printf("ID: %d, ProjectID: %s, Url: %s\n", repository.ID, repository.ProjectID, repository.RootUrl)
		if repository.RootUrl != newUrl {
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
