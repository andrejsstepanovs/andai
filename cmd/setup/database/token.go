package database

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewGetTokenCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Set (or get) redmine admin token",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Get redmine admin token or creates it if missing")

			admin, err := redmine.ApiAdmin()
			if err != nil {
				return fmt.Errorf("error redmine admin: %v", err)
			}
			fmt.Println("Admin ID:", admin.Id)

			token, err := redmine.DbGetToken(admin.Id)
			if err != nil {
				return fmt.Errorf("db err: %v", err)
			}
			if token.ID > 0 {
				fmt.Println("Token already exists")
				fmt.Println("Token:", token.Value)
				return nil
			}

			err = redmine.DbCreateApiToken(admin.Id, viper.GetString("redmine.api_key"))
			if err != nil {
				return fmt.Errorf("after created err: %v", err)
			}
			fmt.Println("New token created")

			token, err = redmine.DbGetToken(admin.Id)
			if err != nil {
				panic(err)
			}
			fmt.Println("Token:", token.Value)

			return nil
		},
	}
	return cmd
}
