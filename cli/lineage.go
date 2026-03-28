package cli

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/client"
	compassv1beta1 "github.com/raystack/compass/proto/raystack/compass/v1beta1"
	"github.com/raystack/salt/cli/printer"
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

			clnt, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}

			req := client.NewRequest(cfg.Client, namespaceID, &compassv1beta1.GetGraphRequest{
				Urn: args[0],
			})
			res, err := clnt.GetGraph(cmd.Context(), req)
			if err != nil {
				return err
			}

			fmt.Println(printer.Bluef("%s", prettyPrint(res.Msg.GetData())))

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&namespaceID, "namespace", "n", namespace.DefaultNamespace.ID.String(), "namespace id or name")
	return cmd
}
