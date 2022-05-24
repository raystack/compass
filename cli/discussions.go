package cli

import (
	"fmt"
	"os"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/internal/client"
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

	cmd.AddCommand(
		listAllDiscussionsCommand(),
		viewDiscussionByIDCommand(),
		postDiscussionCommand(),
	)

	return cmd
}

func listAllDiscussionsCommand() *cobra.Command {
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
			cs := term.NewColorScheme()

			clnt, cancel, err := client.Create(cmd.Context())
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context())
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

				fmt.Println(cs.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			} else {
				fmt.Println(cs.Bluef(prettyPrint(res.GetData())))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&json, "out", "o", "table", "flag to control output viewing, for json `-o json`")

	return cmd
}

func viewDiscussionByIDCommand() *cobra.Command {
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

			clnt, cancel, err := client.Create(cmd.Context())
			if err != nil {
				return err
			}
			defer cancel()

			discussionID := args[0]
			ctx := client.SetMetadata(cmd.Context())
			res, err := clnt.GetDiscussion(ctx, &compassv1beta1.GetDiscussionRequest{
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

func postDiscussionCommand() *cobra.Command {
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

			clnt, cancel, err := client.Create(cmd.Context())
			if err != nil {
				return err
			}
			defer cancel()

			ctx := client.SetMetadata(cmd.Context())
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
