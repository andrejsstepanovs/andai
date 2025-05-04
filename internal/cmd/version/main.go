package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

const ver = "v0.9"

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "This command dose nothing",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("AndAI %s by Andrejs Stepanovs and contributors\n", ver)
		},
	}

	return cmd
}
