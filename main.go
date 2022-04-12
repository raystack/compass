package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/columbus/cmd"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	var command = &cobra.Command{
		Use:           "columbus <command>",
		Short:         "Discovery & Lineage Service",
		Long:          "Metadata Discovery & Lineage Service.",
		SilenceErrors: true,
		SilenceUsage:  false,
		Example: heredoc.Doc(`
			$ columbus serve
			$ columbus migrate
		`),
		Annotations: map[string]string{
			"group:core": "true",
			"help:learn": heredoc.Doc(`
				Use 'columbus <command> <subcommand> --help' for more information about a command.
				Read the manual at https://odpf.github.io/columbus/
			`),
			"help:feedback": heredoc.Doc(`
				Open an issue here https://github.com/odpf/columbus/issues
			`),
		},
	}

	cmdx.SetHelp(command)
	command.AddCommand(serveCmd())
	command.AddCommand(migrateCmd())

	if err := command.Execute(); err != nil {
		if strings.HasPrefix(err.Error(), "unknown command") {
			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Println()
			}
			fmt.Println(command.UsageString())
			os.Exit(exitOK)
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitError)
		}
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve HTTP service",
		Long:  heredoc.Doc(`Serve a HTTP service on a port defined in PORT env var.`),
		Example: heredoc.Doc(`
			$ columbus serve
		`),
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(command *cobra.Command, args []string) error {
			return cmd.Serve()
		},
	}
}

func migrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run storage migration",
		Example: heredoc.Doc(`
			$ columbus migrate
		`),
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(command *cobra.Command, args []string) error {
			cmd.RunMigrate()

			return nil
		},
	}
}
