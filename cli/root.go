package cli

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/compass/internal/client"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

func New(cfg *Config) *cobra.Command {
	cliConfig = cfg

	var rootCmd = &cobra.Command{
		Use:           "compass <command> <subcommand> [flags]",
		Short:         "Discovery & Lineage Service",
		Long:          "Metadata Discovery & Lineage Service.",
		SilenceErrors: true,
		SilenceUsage:  false,
		Example: heredoc.Doc(`
		$ compass asset
		$ compass discussion
		$ compass search
		$ compass server
		`),
		Annotations: map[string]string{
			"group": "core",
			"help:learn": heredoc.Doc(`
				Use 'compass <command> --help' for info about a command.
				Read the manual at https://odpf.github.io/compass/
			`),
			"help:feedback": heredoc.Doc(`
				Open an issue here https://github.com/odpf/compass/issues
			`),
		},
	}

	client.SetConfig(cliConfig.Client)

	rootCmd.AddCommand(
		serverCmd(),
		configCommand(),
		assetsCommand(),
		discussionsCommand(),
		searchCommand(),
		lineageCommand(),
		versionCmd(),
	)

	// Help topics
	rootCmd.AddCommand(cmdx.SetCompletionCmd("compass"))
	rootCmd.AddCommand(cmdx.SetRefCmd(rootCmd))
	rootCmd.AddCommand(cmdx.SetHelpTopicCmd("environment", envHelp))
	cmdx.SetHelp(rootCmd)

	if cliConfig.Client.ServerHeaderKeyUUID == "" {
		cliConfig.Client.ServerHeaderKeyUUID = cliConfig.Service.Identity.HeaderKeyUUID
	}

	rootCmd.PersistentFlags().StringP(configFlag, "c", "", "Override config file")

	return rootCmd
}
