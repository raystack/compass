package cli

import (
	"fmt"
	"os"

	"github.com/odpf/compass/internal/client"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/term"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

const (
	pageSize   = 10
	pageOffset = 0
)

func assetsCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "asset",
		Aliases: []string{"assets"},
		Short:   "Manage assets",
		Annotations: map[string]string{
			"group": "core",
		},
		Example: heredoc.Doc(`
		$ compass asset list
		$ compass asset view
		$ compass asset delete
		$ compass asset edit
		$ compass asset types
		$ compass asset star <id>
		$ compass asset unstar <id>
		$ compass asset starred
		$ compass asset stargazers <id>
		$ compass asset versionhistory <id>
		$ compass asset version <id> <version>
		`),
	}

	cmd.AddCommand(
		listAllAssetsCommand(cfg),
		viewAssetByIDCommand(cfg),
		editAssetCommand(cfg),
		deleteAssetByIDCommand(cfg),
		listAllTypesCommand(cfg),
		listAssetStargazerCommand(cfg),
		starAssetCommand(cfg),
		unstarAssetCommand(cfg),
		starredAssetCommand(cfg),
		versionHistoryAssetCommand(cfg),
		viewAssetByVersionCommand(cfg),
	)

	return cmd
}

func listAllAssetsCommand(cfg *Config) *cobra.Command {
	var types, services, q, qFields, sort, sort_dir, output string
	var data map[string]string
	var size, page uint32
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all assets",
		Example: heredoc.Doc(`
			$ compass asset list
		`),
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.GetAllAssets(ctx, &compassv1beta1.GetAllAssetsRequest{
				Q:         q,
				QFields:   qFields,
				Types:     types,
				Services:  services,
				Sort:      sort,
				Direction: sort_dir,
				Data:      data,
				Size:      size,
				Offset:    page,
			})

			if err != nil {
				return err
			}

			spinner.Stop()
			if output != "json" {
				report := [][]string{}
				report = append(report, []string{"ID", "TYPE", "SERVICE", "URN", "NAME", "VERSION"})
				for _, i := range res.GetData() {
					report = append(report, []string{i.Id, i.Type, i.Service, i.Urn, term.Bluef(i.Name), i.Version})
				}
				printer.Table(os.Stdout, report)

				fmt.Println(term.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			} else {
				fmt.Println(term.Bluef(prettyPrint(res.GetData())))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&types, "types", "t", "", "filter by types")
	cmd.Flags().StringVarP(&services, "services", "s", "", "filter by services")
	cmd.Flags().StringToStringVarP(&data, "data", "d", nil, "filter by field in asset.data")
	cmd.Flags().StringVar(&q, "query", "", "querying by field")
	cmd.Flags().StringVar(&qFields, "query_fields", "", "querying by fields")
	cmd.Flags().StringVar(&sort, "sort", "", "sort by certain fields")
	cmd.Flags().StringVar(&sort_dir, "sort_dir", "", "sorting direction (asc / desc)")
	cmd.Flags().StringVarP(&output, "out", "o", "table", "flag to control output viewing, for json `-o json`")
	cmd.Flags().Uint32Var(&size, "size", pageSize, "Size of each page")
	cmd.Flags().Uint32Var(&page, "page", pageOffset, "Page number offset (starts from 0)")

	return cmd
}

func viewAssetByIDCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "view asset for the given ID",
		Example: heredoc.Doc(`
			$ compass asset view <id>
		`),
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.GetAssetByID(ctx, &compassv1beta1.GetAssetByIDRequest{
				Id: assetID,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(term.Bluef(prettyPrint(res.GetData())))
			return nil
		},
	}

	return cmd
}

func editAssetCommand(cfg *Config) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "upsert a new asset or patch",
		Example: heredoc.Doc(`
			$ compass asset edit --body=filePath
		`),
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody compassv1beta1.UpsertPatchAssetRequest
			if err := parseFile(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.UpsertPatchAsset(ctx, &compassv1beta1.UpsertPatchAssetRequest{
				Asset:     reqBody.Asset,
				Upstreams: reqBody.Upstreams,
			})

			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println("ID: \t", term.Greenf(res.Id))
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "body", "b", "", "filepath to body that has to be upserted")
	if err := cmd.MarkFlagRequired("body"); err != nil {
		panic(err)
	}

	return cmd
}

func deleteAssetByIDCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "delete asset with the given ID",
		Example: heredoc.Doc(`
			$ compass asset delete <id> 
		`),
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			_, err = clnt.DeleteAsset(ctx, &compassv1beta1.DeleteAssetRequest{
				Id: assetID,
			})
			if err != nil {
				return err
			}
			spinner.Stop()
			fmt.Println("Asset ", term.Redf(assetID), " Deleted Successfully")
			return nil
		},
	}

	return cmd
}

func listAllTypesCommand(cfg *Config) *cobra.Command {
	var types, services, q, qFields string
	var data map[string]string

	cmd := &cobra.Command{
		Use:   "types",
		Short: "lists all asset types",
		Example: heredoc.Doc(`
			$ compass asset types
			$ compass asset types -t type1 --query query1 --query_fields qFields1
		`),
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.GetAllTypes(ctx, &compassv1beta1.GetAllTypesRequest{
				Q:        q,
				QFields:  qFields,
				Types:    types,
				Services: services,
				Data:     data,
			})

			if err != nil {
				return err
			}
			spinner.Stop()

			report := [][]string{{"NAME", "COUNT"}}
			for _, i := range res.GetData() {
				report = append(report, []string{term.Bluef(i.Name), fmt.Sprintf("%v", i.Count)})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}
	cmd.Flags().StringVarP(&types, "types", "t", "", "filter by types")
	cmd.Flags().StringVarP(&services, "services", "s", "", "filter by services")
	cmd.Flags().StringToStringVarP(&data, "data", "d", nil, "filter by field in asset.data")
	cmd.Flags().StringVar(&q, "query", "", "filter by specific query")
	cmd.Flags().StringVar(&qFields, "query_fields", "", "filter by query field")

	return cmd
}

func listAssetStargazerCommand(cfg *Config) *cobra.Command {
	var size, page uint32
	cmd := &cobra.Command{
		Use:   "stargazers <id>",
		Short: "list all stargazers for a given asset id",
		Example: heredoc.Doc(`
			$ compass asset stargazers <id>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.GetAssetStargazers(ctx, &compassv1beta1.GetAssetStargazersRequest{
				Id:     assetID,
				Size:   size,
				Offset: page,
			})
			if err != nil {
				return err
			}
			spinner.Stop()
			fmt.Println(term.Bluef(prettyPrint(res.GetData())))
			return nil
		},
	}
	cmd.Flags().Uint32Var(&size, "size", pageSize, "Size of each page")
	cmd.Flags().Uint32Var(&page, "page", pageOffset, "Page number offset (starts from 0)")

	return cmd
}

func starAssetCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "star <id>",
		Short: "star an asset by id for current user",
		Example: heredoc.Doc(`
			$ compass asset star <id>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			_, err = clnt.StarAsset(ctx, &compassv1beta1.StarAssetRequest{
				AssetId: assetID,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(term.Bluef("Asset %v starred successfully", assetID))

			return nil
		},
	}

	return cmd
}

func unstarAssetCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstar <id>",
		Short: "unstar an asset by id for current user",
		Example: heredoc.Doc(`
			$ compass unasset star <id>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			_, err = clnt.UnstarAsset(ctx, &compassv1beta1.UnstarAssetRequest{
				AssetId: assetID,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(term.Bluef("Asset %v unstarred successfully", assetID))

			return nil
		},
	}

	return cmd
}

func starredAssetCommand(cfg *Config) *cobra.Command {
	var size, page uint32
	var output string
	cmd := &cobra.Command{
		Use:   "starred",
		Short: "list all the starred assets for current user",
		Example: heredoc.Doc(`
			$ compass asset starred
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.GetMyStarredAssets(ctx, &compassv1beta1.GetMyStarredAssetsRequest{
				Size:   size,
				Offset: page,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			if output != "json" {
				report := [][]string{}
				report = append(report, []string{"ID", "TYPE", "SERVICE", "URN", "NAME", "VERSION"})
				for _, i := range res.GetData() {
					report = append(report, []string{i.Id, i.Type, i.Service, i.Urn, term.Bluef(i.Name), i.Version})
				}
				printer.Table(os.Stdout, report)

				fmt.Println(term.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			} else {
				fmt.Println(term.Bluef(prettyPrint(res.GetData())))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "out", "o", "table", "flag to control output viewing, for json `-o json`")
	cmd.Flags().Uint32Var(&size, "size", pageSize, "Size of each page")
	cmd.Flags().Uint32Var(&page, "page", pageOffset, "Page number offset (starts from 0)")
	return cmd
}

func versionHistoryAssetCommand(cfg *Config) *cobra.Command {
	var size, page uint32
	cmd := &cobra.Command{
		Use:   "versionhistory <id>",
		Short: "get asset version history by id",
		Example: heredoc.Doc(`
			$ compass asset versionhistory <id>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.GetAssetVersionHistory(ctx, &compassv1beta1.GetAssetVersionHistoryRequest{
				Id:     assetID,
				Size:   size,
				Offset: page,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(term.Bluef(prettyPrint(res.GetData())))

			return nil
		},
	}

	cmd.Flags().Uint32Var(&size, "size", pageSize, "Size of each page")
	cmd.Flags().Uint32Var(&page, "page", pageOffset, "Page number offset (start from 0)")

	return cmd
}

func viewAssetByVersionCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version <id> <version>",
		Short: "get asset's previous version by id and version number",
		Example: heredoc.Doc(`
			$ compass asset version <id> <version>
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			assetID := args[0]
			assetVersion := args[1]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client)
			res, err := clnt.GetAssetByVersion(ctx, &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: assetVersion,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(term.Bluef(prettyPrint(res.GetData())))

			return nil
		},
	}

	return cmd
}
