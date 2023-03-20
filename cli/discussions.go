package cli

import (
	"fmt"
	"github.com/odpf/compass/core/namespace"
	"os"

	"github.com/odpf/compass/internal/client"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func discussionsCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "discussion",
		Aliases: []string{"discussions"},
		Short:   "Manage discussions",
		Annotations: map[string]string{
			"group": "core",
		},
		Example: heredoc.Doc(`
			$ compass discussion list
			$ compass discussion view
			$ compass discussion post
		`),
	}

	cmd.AddCommand(
		listAllDiscussionsCommand(cfg),
		viewDiscussionByIDCommand(cfg),
		postDiscussionCommand(cfg),
	)
	cmd.PersistentFlags().StringVarP(&namespaceID, "namespace", "n", namespace.DefaultNamespace.ID.String(), "namespace id or name")
	return cmd
}

func listAllDiscussionsCommand(cfg *Config) *cobra.Command {
	var json string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all discussions",
		Example: heredoc.Doc(`
			$ compass discussion list
		`),
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

			ctx := client.SetMetadata(cmd.Context(), cfg.Client, namespaceID)
			res, err := clnt.GetAllDiscussions(ctx, &compassv1beta1.GetAllDiscussionsRequest{})
			if err != nil {
				return err
			}

			if json != "json" {
				report := [][]string{}
				report = append(report, []string{"ID", "TITLE", "TYPE", "STATE"})
				index := 1
				for _, i := range res.GetData() {
					report = append(report, []string{i.Id, i.Title, i.Type, i.State})
					index++
				}
				printer.Table(os.Stdout, report)

				fmt.Println(term.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			} else {
				fmt.Println(term.Bluef(prettyPrint(res.GetData())))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&json, "out", "o", "table", "flag to control output viewing, for json `-o json`")

	return cmd
}

func viewDiscussionByIDCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "view discussion for the given ID",
		Example: heredoc.Doc(`
			$ compass discussion view <id>
		`),
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

			discussionID := args[0]
			ctx := client.SetMetadata(cmd.Context(), cfg.Client, namespaceID)
			res, err := clnt.GetDiscussion(ctx, &compassv1beta1.GetDiscussionRequest{
				Id: discussionID,
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

func postDiscussionCommand(cfg *Config) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "post",
		Short: "post discussions, add ",
		Example: heredoc.Doc(`
			$ compass discussion post
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody compassv1beta1.Discussion
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

			ctx := client.SetMetadata(cmd.Context(), cfg.Client, namespaceID)
			res, err := clnt.CreateDiscussion(ctx, &compassv1beta1.CreateDiscussionRequest{
				Title:  reqBody.Title,
				Body:   reqBody.Body,
				Type:   reqBody.Type,
				State:  reqBody.State,
				Labels: reqBody.Labels,
				Assets: reqBody.Assets,
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
