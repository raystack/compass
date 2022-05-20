package cmd

import (
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func discussionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "discussion",
		Aliases: []string{"discussions"},
		Short:   "Manage discussions",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Example: heredoc.Doc(`
			$ compass discussion list
			$ compass discussion view
			$ compass discussion post
		`),
	}

	cmd.AddCommand(listAllDiscussionsCommand())
	cmd.AddCommand(viewdiscussionByIDCommand())
	cmd.AddCommand(postdiscussionCommand())

	return cmd
}

func listAllDiscussionsCommand() *cobra.Command {
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
			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd.Context(), host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context())
			res, err := client.GetAllDiscussions(ctx, &compassv1beta1.GetAllDiscussionsRequest{})
			if err != nil {
				return err
			}

			fmt.Println(cs.Bluef(prettyPrint(res.GetData())))

			return nil
		},
	}

	return cmd
}

func viewdiscussionByIDCommand() *cobra.Command {
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
			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd.Context(), host)
			if err != nil {
				return err
			}
			defer cancel()

			discussionID := args[0]
			ctx := setCtxHeader(cmd.Context())
			res, err := client.GetDiscussion(ctx, &compassv1beta1.GetDiscussionRequest{
				Id: discussionID,
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

func postdiscussionCommand() *cobra.Command {
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
			cs := term.NewColorScheme()

			var reqBody compassv1beta1.Discussion
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
			res, err := client.CreateDiscussion(ctx, &compassv1beta1.CreateDiscussionRequest{
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
