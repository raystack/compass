package cli

import (
	"fmt"
	"github.com/raystack/compass/core/namespace"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/internal/client"
	compassv1beta1 "github.com/raystack/compass/proto/raystack/compass/v1beta1"
	"github.com/raystack/salt/printer"
	"github.com/raystack/salt/term"
	"github.com/spf13/cobra"
)

func lineageCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lineage <urn>",
		Aliases: []string{},
		Short:   "observe the lineage of metadata",
		Annotations: map[string]string{
			"group": "core",
		},
		Args: cobra.ExactArgs(1),
		Example: heredoc.Doc(`
			$ compass lineage <urn>
		`),

		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			clnt, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context(), cfg.Client, namespaceID)

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
	cmd.PersistentFlags().StringVarP(&namespaceID, "namespace", "n", namespace.DefaultNamespace.ID.String(), "namespace id or name")
	return cmd
}
