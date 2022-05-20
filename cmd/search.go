package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
)

var (
	filter, query, rankby string
	size                  uint32
)

func searchCommand() *cobra.Command {
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
			cs := term.NewColorScheme()
			client, cancel, err := createClient(cmd.Context(), host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context())
			res, err := client.SearchAssets(ctx, makeSearchAssetRequest(args[0]))
			if err != nil {
				return err
			}

			fmt.Println(cs.Bluef(prettyPrint(res.GetData())))

			return nil
		},
	}

	cmd.Flags().StringVarP(&filter, "filter", "f", "", "--filter=field_key1:val1,key2:val2,key3:val3 gives exact match for values")
	cmd.Flags().StringVarP(&query, "query", "q", "", "--query=--filter=field_key1:val1 supports fuzzy search")
	cmd.Flags().StringVarP(&rankby, "rankby", "r", "", "--rankby=<numeric_field>")
	cmd.Flags().Uint32VarP(&size, "size", "s", 0, "--size=10 maximum size of response query")
	return cmd
}

func makeSearchAssetRequest(text string) *compassv1beta1.SearchAssetsRequest {
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
