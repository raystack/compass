package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/client"
	"github.com/raystack/compass/internal/config"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
	"github.com/raystack/salt/cli/printer"
	"github.com/spf13/cobra"
)

func namespacesCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"ns"},
		Short:   "Manage namespaces",
		Annotations: map[string]string{
			"group": "core",
		},
		Example: heredoc.Doc(`
			$ compass namespace list
			$ compass namespace view
			$ compass namespace create
		`),
	}

	cmd.AddCommand(
		listNamespacesCommand(cfg),
		getNamespaceCommand(cfg),
		createNamespaceCommand(cfg),
	)
	return cmd
}

func listNamespacesCommand(cfg *config.Config) *cobra.Command {
	var json string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all namespaces",
		Example: heredoc.Doc(`
			$ compass namespace list
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			cl, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}

			req := client.NewRequest(cfg.Client, "", &compassv1beta1.ListNamespacesRequest{})
			res, err := cl.ListNamespaces(cmd.Context(), req)
			if err != nil {
				return err
			}

			if json != "json" {
				var report [][]string
				report = append(report, []string{"ID", "NAME", "STATE"})
				index := 1
				for _, i := range res.Msg.GetNamespaces() {
					report = append(report, []string{i.GetId(), i.GetName(), i.GetState()})
					index++
				}
				printer.Table(os.Stdout, report)

				fmt.Println(printer.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			} else {
				fmt.Println(printer.Bluef("%s", prettyPrint(res.Msg.GetNamespaces())))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&json, "out", "o", "table", "flag to control output viewing, for json `-o json`")
	return cmd
}

func getNamespaceCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "view namespace for the given uuid or name",
		Example: heredoc.Doc(`
			$ compass namespace view <id>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			cl, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}

			urn := args[0]
			req := client.NewRequest(cfg.Client, "", &compassv1beta1.GetNamespaceRequest{
				Urn: urn,
			})
			res, err := cl.GetNamespace(cmd.Context(), req)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(printer.Bluef("%s", prettyPrint(res.Msg.GetNamespace())))
			return nil
		},
	}

	return cmd
}

func createNamespaceCommand(cfg *config.Config) *cobra.Command {
	var name, state string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a new namespace",
		Example: heredoc.Doc(`
			$ compass namespace create
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			if len(name) < 3 || strings.ContainsAny(name, " .-") {
				return errors.New("namespace length should be of at least 3 character, without space and special characters")
			}

			cl, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}

			req := client.NewRequest(cfg.Client, "", &compassv1beta1.CreateNamespaceRequest{
				Name:  name,
				State: state,
			})
			res, err := cl.CreateNamespace(cmd.Context(), req)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println("ID: \t", printer.Greenf("%s", res.Msg.Id))
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "namespace unique name")
	cmd.Flags().StringVarP(&state, "state", "s", namespace.SharedState.String(), "is namespace shared with existing tenants or a dedicated one")
	return cmd
}
