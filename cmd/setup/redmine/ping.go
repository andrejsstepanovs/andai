package redmine

import (
	"fmt"

	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewPingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping connection to redmine",
		RunE:  runRedminePing,
	}
	return cmd
}

func runRedminePing(cmd *cobra.Command, args []string) error {
	fmt.Println("Processing Jira issue", len(args))

	client := redmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))
	projects, err := client.News(1)
	if err != nil {
		fmt.Println("Redmine Ping Fail")
		return fmt.Errorf("error redmine ping: %v", err)
	}
	for _, project := range projects {
		fmt.Println(project)
	}

	fmt.Println("Redmine Ping Success")

	return nil
}
