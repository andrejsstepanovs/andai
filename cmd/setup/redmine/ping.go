package redmine

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewPingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redmine",
		Short: "Ping connection to redmine",
		RunE:  runRedminePing,
	}
	return cmd
}

func runRedminePing(cmd *cobra.Command, args []string) error {
	fmt.Println("Processing Jira issue", len(args))

	client := redmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))
	users, err := client.Users()
	if err != nil {
		fmt.Println("Redmine Ping Fail")
		return fmt.Errorf("error redmine ping: %v", err)
	}
	for _, user := range users {
		fmt.Println(fmt.Sprintf("ID: %d, Name: %s", user.Id, user.Login))
	}
	fmt.Println("Redmine Ping Success")
	return nil
}
