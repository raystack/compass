package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Println(Version)
			return err
		},
	}
}
