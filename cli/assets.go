package cli

import (
	"fmt"
	"os"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/term"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
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
			$ compass asset view
			$ compass asset delete
			$ compass asset edit
		`),
	}

	cmd.AddCommand(listAllAssetsCommand())
	cmd.AddCommand(viewAssetByIDCommand())
	cmd.AddCommand(editAssetCommand())
	cmd.AddCommand(deleteAssetByIDCommand())

	return cmd
}

func listAllAssetsCommand() *cobra.Command {
	var types, services, data, q, sort, sort_dir, json string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all assets",
		Example: heredoc.Doc(`
			$ compass asset list
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()
			client, cancel, err := createClient(cmd.Context(), host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context())
			res, err := client.GetAllAssets(ctx, makeGetAllAssetRequest(types, services, data, q, sort, sort_dir))
			if err != nil {
				return err
			}
			if json != "json" {
				report := [][]string{}
				report = append(report, []string{"ID", "TYPE", "SERVICE", "URN", "NAME", "VERSION"})
				index := 1
				for _, i := range res.GetData() {
					report = append(report, []string{i.Id, i.Type, i.Service, i.Urn, cs.Bluef(i.Name), i.Version})
					index++
				}
				printer.Table(os.Stdout, report)

				fmt.Println(cs.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			} else {
				fmt.Println(cs.Bluef(prettyPrint(res.GetData())))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&types, "types", "t", "", "filter by types")
	cmd.Flags().StringVarP(&services, "services", "s", "", "filter by services")
	cmd.Flags().StringVarP(&data, "data", "d", "", "filter by field in asset.data")
	cmd.Flags().StringVar(&q, "query", "", "querying by field")
	cmd.Flags().StringVar(&sort, "sort", "", "sort by certain fields")
	cmd.Flags().StringVar(&sort_dir, "sort_dir", "", "sorting direction (asc / desc)")
	cmd.Flags().StringVarP(&json, "out", "o", "table", "flag to control output viewing, for json `-o json`")

	return cmd
}

func makeGetAllAssetRequest(types, services, data, q, sort, sort_dir string) *compassv1beta1.GetAllAssetsRequest {
	newReq := &compassv1beta1.GetAllAssetsRequest{}
	if types != "" {
		newReq.Types = types
	}
	if services != "" {
		newReq.Services = services
	}
	if q != "" {
		newReq.Q = q
	}
	if sort != "" {
		newReq.Sort = sort
	}
	if sort_dir != "" {
		newReq.Direction = sort_dir
	}
	if data != "" {
		newReq.Data = makeMapFromString(data)
	}

	return newReq
}

func viewAssetByIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "view asset for the given ID",
		Example: heredoc.Doc(`
			$ compass asset view <id>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd.Context(), host)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := setCtxHeader(cmd.Context())
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

func editAssetCommand() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "upsert a new asset or patch",
		Example: heredoc.Doc(`
			$ compass asset edit --body=filePath
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody compassv1beta1.UpsertPatchAssetRequest
			if err := parseFile(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd.Context(), host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context())

			res, err := client.UpsertPatchAsset(ctx, &compassv1beta1.UpsertPatchAssetRequest{
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
		Short: "delete asset with the given ID",
		Example: heredoc.Doc(`
			$ compass asset delete <id> 
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd.Context(), host)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := setCtxHeader(cmd.Context())
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
