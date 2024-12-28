package setup

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newGetTokenCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Set (or get) redmine admin token",
		RunE: func(cmd *cobra.Command, args []string) error {
			newToken := viper.GetString("redmine.api_key")
			fmt.Println("Get redmine admin token or creates it if missing")

			admin, err := redmine.ApiAdmin()
			if err != nil {
				return fmt.Errorf("error redmine admin: %v", err)
			}
			fmt.Println("Admin ID:", admin.Id)

			getToken := func() (models.Token, error) {
				token, err := redmine.DbGetToken(admin.Id)
				if err != nil {
					return models.Token{}, fmt.Errorf("db err: %v", err)
				}
				if token.ID > 0 {
					fmt.Println("Token:", token.Value)
				}
				return token, nil
			}

			token, err := getToken()
			if err != nil {
				return err
			}

			if token.ID > 0 {
				if token.Value == newToken {
					return nil
				}

				fmt.Println("Token mismatch")
				err = redmine.DbUpdateApiToken(admin.Id, newToken)
				if err != nil {
					return fmt.Errorf("after updated err: %v", err)
				}
				fmt.Println("Token updated")
				token, err = getToken()
				if err != nil {
					return err
				}
				return nil
			}

			err = redmine.DbCreateApiToken(admin.Id, newToken)
			if err != nil {
				return fmt.Errorf("after created err: %v", err)
			}
			fmt.Println("New token created")

			_, err = getToken()
			return err
		},
	}
	return cmd
}
