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
		Use:     "discussions",
		Aliases: []string{"discussions"},
		Short:   "Manage discussions",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Example: heredoc.Doc(`
			$ compass discussions list
			$ compass discussions get
			$ compass discussions post
		`),
	}

	cmd.AddCommand(listAllDiscussionsCommand())
	cmd.AddCommand(getdiscussionByIDCommand())
	cmd.AddCommand(postdiscussionCommand())

	return cmd
}

func listAllDiscussionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all discussions",
		Example: heredoc.Doc(`
			$ compass discussions list --host=<hostaddress> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd, host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)
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

func getdiscussionByIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "get discussions for the given ID",
		Example: heredoc.Doc(`
			$ compass discussions get <id> --host=<hostaddress> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd, host)
			if err != nil {
				return err
			}
			defer cancel()

			discussionID := args[0]
			ctx := setCtxHeader(cmd.Context(), header)
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
			$ compass discussions post --host=<hostaddress> --header=<key>:<value> --body=filePath
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

			client, cancel, err := createClient(cmd, host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)
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
	cmd.MarkFlagRequired("body")

	return cmd
}
