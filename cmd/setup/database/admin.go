package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Fix admin",
		RunE:  runDatabaseAdmin,
	}
	return cmd
}

func runDatabaseAdmin(cmd *cobra.Command, args []string) error {
	fmt.Println("Update redmine admin must_change_passwd = 0")

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		panic(err)
	}

	result, err := db.Exec("UPDATE users SET must_change_passwd = 0 WHERE login = ?", "admin")
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
