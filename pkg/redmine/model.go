package redmine

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

const ADMIN_LOGIN = "admin"

type Model struct {
	db  *sql.DB
	api *redmine.Client
}

func NewModel() *Model {
	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		panic(err)
	}
	return &Model{
		api: redmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key")),
		db:  db,
	}
}

func (c *Model) Api() *redmine.Client {
	return c.api
}

func (c *Model) Db() *sql.DB {
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
