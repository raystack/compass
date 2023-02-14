package cli

import (
	"fmt"
	"os"
	"strings"

	client "github.com/odpf/compass/internal/client"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

const (
	exitOK    = 0
	exitError = 1
)

var rootCmd = &cobra.Command{
	Use:           "compass <command> <subcommand>",
	Short:         "Discovery & Lineage Service",
	Long:          "Metadata Discovery & Lineage Service.",
	SilenceErrors: true,
	SilenceUsage:  false,
	Example: heredoc.Doc(`
		$ compass serve
		$ compass migrate
	`),
	Annotations: map[string]string{
		"group": "core",
		"help:learn": heredoc.Doc(`
			Use 'compass <command> <subcommand> --help' for more information about a command.
			Read the manual at https://odpf.github.io/compass/
		`),
		"help:feedback": heredoc.Doc(`
			Open an issue here https://github.com/odpf/compass/issues
		`),
	},
}

func Execute() {
	config, err := loadConfig(rootCmd)
	if err != nil {
		panic(err)
	}

	// if Client.ServerHeaderKeyUUID is not set, use the value from server config
	if config.Client.ServerHeaderKeyUUID == "" {
		config.Client.ServerHeaderKeyUUID = config.Service.Identity.HeaderKeyUUID
	}
	client.SetConfig(config.Client)

	rootCmd.PersistentFlags().StringP(configFlag, "c", "", "Override config file")
	rootCmd.AddCommand(
		cmdServe(),
		cmdMigrate(),
		cmdShowConfigs(),
		assetsCommand(),
		discussionsCommand(),
		searchCommand(),
		lineageCommand(),
		versionCmd(),
	)

	// Help topics
	rootCmd.AddCommand(cmdx.SetRefCmd(rootCmd))
	cmdx.SetHelp(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		if strings.HasPrefix(err.Error(), "unknown command") {
			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Println()
			}
			fmt.Println(rootCmd.UsageString())
			os.Exit(exitOK)
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitError)
		}
	}
}
