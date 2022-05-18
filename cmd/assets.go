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
	cmd.AddCommand(getAssetByIDCommand())
	cmd.AddCommand(postAssetCommand())

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

			fmt.Println(prettyPrint(res.GetData()))

			return nil
		},
	}

	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")
	cmd.Flags().StringVarP(&host, "host", "h", "", "Compass service to connect to")
	cmd.MarkFlagRequired("host")

	return cmd
}

func getAssetByIDCommand() *cobra.Command {
	var host, header string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "get asset for the given ID",
		Example: heredoc.Doc(`
			$ compass asset get <id> --host=<hostaddress> --header=<key>:<value>
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

			assetID := args[0]
			ctx := setCtxHeader(cmd.Context(), header)
			res, err := client.GetAssetByID(ctx, &compassv1beta1.GetAssetByIDRequest{
				Id: assetID,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(prettyPrint(res.GetData()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")
	cmd.Flags().StringVarP(&host, "host", "h", "", "Compass service to connect to")
	cmd.MarkFlagRequired("host")

	return cmd
}

func postAssetCommand() *cobra.Command {
	var host, header, filePath string

	cmd := &cobra.Command{
		Use:   "post",
		Short: "post asset, add ",
		Example: heredoc.Doc(`
			$ compass asset post --host=<hostaddress> --header=<key>:<value> --body=filePath
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody compassv1beta1.UpsertAssetRequest
			if err := parseFile(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd, host)
			if err != nil {
				return err
			}
			defer cancel()

			// WIP
			ctx := setCtxHeader(cmd.Context(), header)
			res, err := client.UpsertAsset(ctx, &compassv1beta1.UpsertAssetRequest{
				Asset:     reqBody.Asset,
				Upstreams: reqBody.Upstreams,
			})

			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println("ID: \t", res.Id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "body", "b", "", "filepath to body that has to be upserted")
	cmd.MarkFlagRequired("body")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")
	cmd.Flags().StringVarP(&host, "host", "h", "", "Compass service to connect to")
	cmd.MarkFlagRequired("host")

	return cmd
}
