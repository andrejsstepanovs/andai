package redmine

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
)

const (
	queryGetAllUsers = "SELECT id, login, firstname, lastname FROM users"
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
