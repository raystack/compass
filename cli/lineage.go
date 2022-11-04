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

func lineageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lineage <urn>",
		Aliases: []string{},
		Short:   "observe the lineage of metadata",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		Example: heredoc.Doc(`
			$ compass lineage <urn>
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

			res, err := clnt.GetGraph(ctx, &compassv1beta1.GetGraphRequest{
				Urn: args[0],
			})
			if err != nil {
				return err
			}

			fmt.Println(term.Bluef(prettyPrint(res.GetData())))

			return nil
		},
	}
	return cmd
}
