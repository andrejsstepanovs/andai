package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newGetTokenCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Set (or get) redmine admin token",
		RunE: func(_ *cobra.Command, _ []string) error {
			newToken := viper.GetString("redmine.api_key")
			log.Println("Get redmine admin token or creates it if missing")

			admin, err := redmine.APIAdmin()
			if err != nil {
				return fmt.Errorf("error redmine admin: %v", err)
			}
			log.Println("Admin Identifier:", admin.Id)

			getToken := func() (models.Token, error) {
				token, err := redmine.DBGetToken(admin.Id)
				if err != nil {
					return models.Token{}, fmt.Errorf("db err: %v", err)
				}
				if token.ID > 0 {
					log.Println("Token:", token.Value)
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

				log.Println("Token mismatch")
				err = redmine.DBUpdateAPIToken(admin.Id, newToken)
				if err != nil {
					return fmt.Errorf("after updated err: %v", err)
				}
				log.Println("Token updated")
				token, err = getToken()
				if err != nil {
					return err
				}
				return nil
			}

			err = redmine.DBCreateAPIToken(admin.Id, newToken)
			if err != nil {
				return fmt.Errorf("after created err: %v", err)
			}
			log.Println("New token created")

			_, err = getToken()
			return err
		},
	}
	return cmd
}
