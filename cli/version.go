package cli

import (
	"fmt"

	"github.com/raystack/salt/term"
	"github.com/raystack/salt/version"
	"github.com/spf13/cobra"
)

// VersionCmd prints the version of the binary
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if Version == "" {
				fmt.Println(term.Yellow("Version information not available"))
				return nil
			}

			fmt.Printf("compass version %s", Version)
			fmt.Println(term.Yellow(version.UpdateNotice(Version, "raystack/compass")))
			return nil
		},
	}
}
