package cli

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/goto/salt/cmdx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "compass <command> <subcommand> [flags]",
	Short:         "Discovery & Lineage Service",
	Long:          "Metadata Discovery & Lineage Service.",
	SilenceErrors: true,
	SilenceUsage:  true,
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
			Read the manual at https://goto.github.io/compass/
		`),
		"help:feedback": heredoc.Doc(`
			Open an issue here https://github.com/goto/compass/issues
		`),
	},
}

func New(cliConfig *Config) *cobra.Command {
	if cliConfig.Client.ServerHeaderKeyUUID == "" {
		cliConfig.Client.ServerHeaderKeyUUID = cliConfig.Service.Identity.HeaderKeyUUID
	}

	rootCmd.PersistentPreRunE = func(subCmd *cobra.Command, args []string) error {
		cfgFile, _ := subCmd.Flags().GetString(configFlag)
		if cfgFile != "" {
			err := LoadConfigFromFlag(cfgFile, cliConfig)
			if err != nil {
				return err
			}
		}
		return nil
	}

	rootCmd.PersistentFlags().StringP(configFlag, "c", "", "Override config file")

	rootCmd.AddCommand(
		serverCmd(cliConfig),
		workerCmd(cliConfig),
		configCommand(cliConfig),
		assetsCommand(cliConfig),
		discussionsCommand(cliConfig),
		searchCommand(cliConfig),
		lineageCommand(cliConfig),
		versionCmd(),
	)

	// Help topics
	rootCmd.AddCommand(cmdx.SetCompletionCmd("compass"))
	rootCmd.AddCommand(cmdx.SetRefCmd(rootCmd))
	rootCmd.AddCommand(cmdx.SetHelpTopicCmd("environment", envHelp))
	cmdx.SetHelp(rootCmd)

	return rootCmd
}
