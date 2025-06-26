package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal/exec"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newTreePingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Check if tree command is installed",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := pingTree()
			if err != nil {
				return err
			}

			log.Println("tree command Success")
			return nil
		},
	}
	return cmd
}

func pingTree() error {
	if !exec.IsTreeInstalled() {
		return fmt.Errorf("tree is not installed")
	}
	return nil
}
