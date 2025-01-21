package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newAdminCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Fix admin login no need to change password",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Update redmine admin must_change_passwd = 0")

			err := redmine.DBSettingsAdminMustNotChangePassword()
			if err != nil {
				log.Println("Redmine Admin Setting Change Fail")
				return fmt.Errorf("error db: %v", err)
			}
			return nil
		},
	}
	return cmd
}
