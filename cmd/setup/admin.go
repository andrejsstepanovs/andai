package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newAdminCommand(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Fix admin login no need to change password and other settings",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Update redmine admin must_change_passwd = 0")
			return setupAdmin(deps.Model)
		},
	}
	return cmd
}

func setupAdmin(redmine *redmine.Model) error {
	err := redmine.DBSettingsAdminMustNotChangePassword()
	if err != nil {
		return fmt.Errorf("error db: %v", err)
	}
	return nil
}
