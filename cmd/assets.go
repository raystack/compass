package cmd

import (
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func assetsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "assets",
		Aliases: []string{"assets"},
		Short:   "Manage assets",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Example: heredoc.Doc(`
			$ compass assets list
			$ compass assets get
			$ compass assets delete
			$ compass assets post
		`),
	}

	cmd.AddCommand(listAllAssetsCommand())
	cmd.AddCommand(getAssetByIDCommand())
	cmd.AddCommand(postAssetCommand())
	cmd.AddCommand(deleteAssetByIDCommand())

	return cmd
}

func listAllAssetsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all assets",
		Example: heredoc.Doc(`
			$ compass assets list --host=<hostaddress> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()
			fmt.Println("header:", header)
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

			fmt.Println(cs.Bluef(prettyPrint(res.GetData())))

			return nil
		},
	}

	return cmd
}

func getAssetByIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "get assets for the given ID",
		Example: heredoc.Doc(`
			$ compass assets get <id> --host=<hostaddress> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

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

			fmt.Println(cs.Bluef(prettyPrint(res.GetData())))
			return nil
		},
	}

	return cmd
}

func postAssetCommand() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "post",
		Short: "post assets, add ",
		Example: heredoc.Doc(`
			$ compass assets post --host=<hostaddress> --header=<key>:<value> --body=filePath
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

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

			ctx := setCtxHeader(cmd.Context(), header)
			res, err := client.UpsertAsset(ctx, &compassv1beta1.UpsertAssetRequest{
				Asset:     reqBody.Asset,
				Upstreams: reqBody.Upstreams,
			})

			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println("ID: \t", cs.Greenf(res.Id))
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "body", "b", "", "filepath to body that has to be upserted")
	if err := cmd.MarkFlagRequired("body"); err != nil {
		panic(err)
	}

	return cmd
}

func deleteAssetByIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "delete assets with the given ID",
		Example: heredoc.Doc(`
			$ compass assets delete <id> --host=<hostaddress> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd, host)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := setCtxHeader(cmd.Context(), header)
			_, err = client.DeleteAsset(ctx, &compassv1beta1.DeleteAssetRequest{
				Id: assetID,
			})
			if err != nil {
				return err
			}
			spinner.Stop()
			fmt.Println("Asset ", cs.Redf(assetID), " Deleted Successfully")
			return nil
		},
	}

	return cmd
}