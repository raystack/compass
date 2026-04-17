package cli

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/internal/config"
	"github.com/raystack/salt/cli/commander"
	saltconfig "github.com/raystack/salt/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:           "compass <command> <subcommand> [flags]",
		Short:         "Discovery & Lineage Service",
		Long:          "Metadata Discovery & Lineage Service.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Example: heredoc.Doc(`
			$ compass entity
			$ compass search
			$ compass server
	`),
		Annotations: map[string]string{
			"group": "core",
			"help:learn": heredoc.Doc(`
			Use 'compass <command> --help' for info about a command.
			Read the manual at https://raystack.github.io/compass/
		`),
			"help:feedback": heredoc.Doc(`
			Open an issue here https://github.com/raystack/compass/issues
		`),
		},
	}

)

func New(cliConfig *config.Config) *cobra.Command {
	if cliConfig.Client.ServerHeaderKeyUserUUID == "" {
		// defaulting to server defined key
		cliConfig.Client.ServerHeaderKeyUserUUID = cliConfig.Service.Identity.HeaderKeyUserUUID
	}

	rootCmd.PersistentPreRunE = func(subCmd *cobra.Command, args []string) error {
		cfgFile, _ := subCmd.Flags().GetString(configFlag)
		if cfgFile != "" {
			if err := saltconfig.NewLoader(saltconfig.WithFile(cfgFile)).Load(cliConfig); err != nil {
				return err
			}
		}
		return nil
	}

	rootCmd.PersistentFlags().StringP(configFlag, "c", "", "Override config file")

	rootCmd.AddCommand(
		serverCmd(cliConfig),
		configCommand(cliConfig),
		namespacesCommand(cliConfig),
		entitiesCommand(cliConfig),
		documentsCommand(cliConfig),
		embedCommand(cliConfig),
		versionCmd(),
	)

	// Help topics, completion, reference
	mgr := commander.New(rootCmd, commander.WithTopics([]commander.HelpTopic{
		{
			Name:  "environment",
			Short: envHelp["short"],
			Long:  envHelp["long"],
		},
	}))
	mgr.Init()

	return rootCmd
}
