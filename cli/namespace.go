package cli

import (
	"errors"
	"fmt"
	"github.com/raystack/compass/core/namespace"
	"os"
	"strings"

	"github.com/raystack/compass/internal/client"
	compassv1beta1 "github.com/raystack/compass/proto/raystack/compass/v1beta1"
	"github.com/raystack/salt/printer"
	"github.com/raystack/salt/term"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func namespacesCommand(cfg *Config) *cobra.Command {
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

func listNamespacesCommand(cfg *Config) *cobra.Command {
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

			cl, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context(), cfg.Client, "")
			res, err := cl.ListNamespaces(ctx, &compassv1beta1.ListNamespacesRequest{})
			if err != nil {
				return err
			}

			if json != "json" {
				var report [][]string
				report = append(report, []string{"ID", "NAME", "STATE"})
				index := 1
				for _, i := range res.GetNamespaces() {
					report = append(report, []string{i.GetId(), i.GetName(), i.GetState()})
					index++
				}
				printer.Table(os.Stdout, report)

				fmt.Println(term.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			} else {
				fmt.Println(term.Bluef(prettyPrint(res.GetNamespaces())))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&json, "out", "o", "table", "flag to control output viewing, for json `-o json`")
	return cmd
}

func getNamespaceCommand(cfg *Config) *cobra.Command {
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

			cl, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			urn := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client, "")
			res, err := cl.GetNamespace(ctx, &compassv1beta1.GetNamespaceRequest{
				Urn: urn,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(term.Bluef(prettyPrint(res.GetNamespace())))
			return nil
		},
	}

	return cmd
}

func createNamespaceCommand(cfg *Config) *cobra.Command {
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

			cl, cancel, err := client.Create(cmd.Context(), cfg.Client)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context(), cfg.Client, "")
			res, err := cl.CreateNamespace(ctx, &compassv1beta1.CreateNamespaceRequest{
				Name:  name,
				State: state,
			})

			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println("ID: \t", term.Greenf(res.Id))
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "namespace unique name")
	cmd.Flags().StringVarP(&state, "state", "s", namespace.SharedState.String(), "is namespace shared with existing tenants or a dedicated one")
	return cmd
}
