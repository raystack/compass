package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

var host, header string

func New() *cobra.Command {
	var command = &cobra.Command{
		Use:           "compass <command>",
		Short:         "Discovery & Lineage Service",
		Long:          "Metadata Discovery & Lineage Service.",
		SilenceErrors: true,
		SilenceUsage:  false,
		Example: heredoc.Doc(`
			$ compass serve
			$ compass migrate
		`),
		Annotations: map[string]string{
			"group:core": "true",
			"help:learn": heredoc.Doc(`
				Use 'compass <command> <subcommand> --help' for more information about a command.
				Read the manual at https://odpf.github.io/compass/
			`),
			"help:feedback": heredoc.Doc(`
				Open an issue here https://github.com/odpf/compass/issues
			`),
		},
	}

	if err := loadConfig(); err != nil {
		panic(err)
	}

	host = config.ServerBaseUrl
	header = config.AuthHeader

	cmdx.SetHelp(command)
	command.AddCommand(serveCmd())
	command.AddCommand(migrateCmd())
	command.AddCommand(assetsCommand())
	command.AddCommand(discussionsCommand())
	return command
}
