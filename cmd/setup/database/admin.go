package database

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

func NewAdminCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Fix admin login no need to change password",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Update redmine admin must_change_passwd = 0")

			err := redmine.DbSettingsAdminMustNotChangePassword()
			if err != nil {
				fmt.Println("Redmine Admin Setting Change Fail")
				return fmt.Errorf("error db: %v", err)
			}
			return nil
		},
	}
	return cmd
}
