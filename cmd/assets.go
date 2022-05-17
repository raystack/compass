package cmd

import (
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/printer"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func assetsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "asset",
		Aliases: []string{"assets"},
		Short:   "Manage assets",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Example: heredoc.Doc(`
			$ compass asset list
			$ compass asset get
			$ compass asset delete
			$ compass asset post
		`),
	}

	cmd.AddCommand(listAllAssetsCommand())
	return cmd
}

func listAllAssetsCommand() *cobra.Command {
	var host, header string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all assets",
		Example: heredoc.Doc(`
			$ compass asset list --host=<hostaddress> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd, host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)
			res, err := client.GetAllAssets(ctx, &compassv1beta1.GetAllAssetsRequest{})
			if err != nil {
				return err
			}

			fmt.Println(res.GetData())

			return nil
		},
	}

	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")
	cmd.Flags().StringVarP(&host, "host", "h", "", "Guardian service to connect to")
	cmd.MarkFlagRequired("host")

	return cmd
}
