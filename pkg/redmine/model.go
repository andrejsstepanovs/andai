package redmine

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

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
