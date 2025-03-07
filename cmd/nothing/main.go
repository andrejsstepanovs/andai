package nothing

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nothing",
		Short: "This command dose nothing",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "execute",
			Short: "Sleeps for a year",
			RunE: func(_ *cobra.Command, _ []string) error {
				fmt.Println("UP")
				time.Sleep(time.Hour * 24 * 365)
				return nil
			},
		},
	)

	return cmd
}
