package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSettingsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Setup redmine settings",
		RunE:  runRedmineSettings,
	}
	return cmd
}

func runRedmineSettings(cmd *cobra.Command, args []string) error {
	fmt.Println("Update Redmine settings")

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		panic(err)
	}

	// INSERT INTO `settings` (`id`, `name`, `value`, `updated_on`) VALUES
	//(1, 'default_notification_option', 'only_assigned', NULL),
	//(2, 'text_formatting', 'common_mark', NULL),
	//(3, 'rest_api_enabled', '1', '2024-12-28 19:03:46'),
	//(4, 'jsonp_enabled', '0', '2024-12-28 19:03:46');

	results, err := db.Query("SELECT id, name, value FROM settings WHERE name = ?", "rest_api_enabled")
	if err != nil {
		fmt.Println("Redmine Admin Fail")
		return fmt.Errorf("error redmine admin: %v", err)
	}
	defer results.Close()

	type Row struct {
		ID    int
		Name  string
		Value string
	}
	var rows []Row

	// Loop through rows, using Scan to assign column data to struct fields.
	for results.Next() {
		var row Row
		if err := results.Scan(&row.ID, &row.Name, &row.Value); err != nil {
			return err
		}
		rows = append(rows, row)
	}
	if err = results.Err(); err != nil {
		return err
	}

	if len(rows) > 0 {
		for _, row := range rows {
			fmt.Sprintf("ID: %d, Name: %s, Value: %s", row.ID, row.Name, row.Value)

			if row.Value == "0" {
				fmt.Println("Admin rest_api_enabled is not enabled. Enabling it now")

				result, err := db.Exec("UPDATE settings SET value = ?, updated_on = NOW() WHERE name = ?", "1", "rest_api_enabled")
				if err != nil {
					return fmt.Errorf("update settings db err: %v", err)
				}
				affected, err := result.RowsAffected()
				if err != nil {
					return fmt.Errorf("rows affected err: %v", err)
				}
				if affected > 0 {
					fmt.Println("Admin rest_api_enabled set to 1")
					return nil
				}
				return errors.New("Admin rest_api_enabled not changed")
			}
		}
		fmt.Println("Nothing to change")
		return nil
	}

	result, err := db.Exec("INSERT INTO settings (name, value, updated_on) VALUES (?, ?, NOW())", "rest_api_enabled", "1")
	if err != nil {
		return fmt.Errorf("update settings db err: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}

	if affected > 0 {
		fmt.Println("Admin rest_api_enabled set to 1")
	} else {
		return errors.New("Admin rest_api_enabled not changed")
	}

	return nil
}
