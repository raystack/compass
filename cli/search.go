package cli

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/compass/internal/client"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
)

func searchCommand() *cobra.Command {
	var filter, query, rankby string
	var size uint32
	cmd := &cobra.Command{
		Use:     "search <text>",
		Aliases: []string{},
		Short:   "query the metadata available",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		Example: heredoc.Doc(`
			$ compass search view
		`),

		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			clnt, cancel, err := client.Create(cmd.Context())
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context())
			res, err := clnt.SearchAssets(ctx, makeSearchAssetRequest(args[0], filter, query, rankby, size))
			if err != nil {
				return err
			}

			fmt.Println(term.Bluef(prettyPrint(res.GetData())))

			return nil
		},
	}

	cmd.Flags().StringVarP(&filter, "filter", "f", "", "--filter=field_key1:val1,key2:val2,key3:val3 gives exact match for values")
	cmd.Flags().StringVarP(&query, "query", "q", "", "--query=--filter=field_key1:val1 supports fuzzy search")
	cmd.Flags().StringVarP(&rankby, "rankby", "r", "", "--rankby=<numeric_field>")
	cmd.Flags().Uint32VarP(&size, "size", "s", 0, "--size=10 maximum size of response query")
	return cmd
}

func makeSearchAssetRequest(text, filter, query, rankby string, size uint32) *compassv1beta1.SearchAssetsRequest {
	newReq := &compassv1beta1.SearchAssetsRequest{
		Text: text,
	}
	if filter != "" {
		newReq.Filter = makeMapFromString(filter)
	}
	if query != "" {
		newReq.Query = makeMapFromString(query)
	}
	if rankby != "" {
		newReq.Rankby = rankby
	}
	if size > 0 {
		newReq.Size = size
	}
	return newReq
}
