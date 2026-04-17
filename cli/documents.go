package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/internal/config"
	"github.com/spf13/cobra"
)

func documentsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "document",
		Aliases: []string{"documents", "doc", "docs"},
		Short:   "Manage documents in the knowledge graph",
		Annotations: map[string]string{
			"group": "core",
		},
		Example: heredoc.Doc(`
		$ compass document list
		$ compass document view <id>
		$ compass document upsert
		$ compass document delete <id>
		$ compass document entity <urn>
		`),
	}

	cmd.AddCommand(
		listDocumentsCommand(cfg),
		viewDocumentCommand(cfg),
		upsertDocumentCommand(cfg),
		deleteDocumentCommand(cfg),
		documentsByEntityCommand(cfg),
	)

	return cmd
}

func listDocumentsCommand(cfg *config.Config) *cobra.Command {
	var entityURN, source string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("http://%s/v1/documents", cfg.Client.Host)
			params := ""
			if entityURN != "" {
				params += "?entity_urn=" + entityURN
			}
			if source != "" {
				sep := "?"
				if params != "" {
					sep = "&"
				}
				params += sep + "source=" + source
			}

			body, err := doDocumentRequest(cfg, "GET", url+params, nil)
			if err != nil {
				return err
			}

			fmt.Println(string(body))
			return nil
		},
	}
	cmd.Flags().StringVar(&entityURN, "entity-urn", "", "Filter by entity URN")
	cmd.Flags().StringVar(&source, "source", "", "Filter by source")
	return cmd
}

func viewDocumentCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View document by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("http://%s/v1/documents/%s", cfg.Client.Host, args[0])

			body, err := doDocumentRequest(cfg, "GET", url, nil)
			if err != nil {
				return err
			}

			fmt.Println(string(body))
			return nil
		},
	}
}

func upsertDocumentCommand(cfg *config.Config) *cobra.Command {
	var entityURN, title, docBody, format, source, sourceID string

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update a document",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{
				"entity_urn": entityURN,
				"title":      title,
				"body":       docBody,
			}
			if format != "" {
				payload["format"] = format
			}
			if source != "" {
				payload["source"] = source
			}
			if sourceID != "" {
				payload["source_id"] = sourceID
			}

			url := fmt.Sprintf("http://%s/v1/documents", cfg.Client.Host)

			body, err := doDocumentRequest(cfg, "POST", url, payload)
			if err != nil {
				return err
			}

			var res map[string]string
			if err := json.Unmarshal(body, &res); err == nil {
				if id, ok := res["id"]; ok {
					fmt.Println("Document upserted:", id)
					return nil
				}
			}
			fmt.Println(string(body))
			return nil
		},
	}
	cmd.Flags().StringVar(&entityURN, "entity-urn", "", "Entity URN (required)")
	cmd.Flags().StringVar(&title, "title", "", "Document title (required)")
	cmd.Flags().StringVar(&docBody, "body", "", "Document body (required)")
	cmd.Flags().StringVar(&format, "format", "", "Format: markdown, plaintext")
	cmd.Flags().StringVar(&source, "source", "", "Source system (e.g., confluence, github)")
	cmd.Flags().StringVar(&sourceID, "source-id", "", "ID in source system")
	_ = cmd.MarkFlagRequired("entity-urn")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

func deleteDocumentCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a document by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("http://%s/v1/documents/%s", cfg.Client.Host, args[0])

			if _, err := doDocumentRequest(cfg, "DELETE", url, nil); err != nil {
				return err
			}

			fmt.Println("Document deleted:", args[0])
			return nil
		},
	}
}

func documentsByEntityCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "entity <urn>",
		Short: "List documents for an entity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("http://%s/v1/entities/%s/documents", cfg.Client.Host, args[0])

			body, err := doDocumentRequest(cfg, "GET", url, nil)
			if err != nil {
				return err
			}

			fmt.Println(string(body))
			return nil
		},
	}
}

func doDocumentRequest(cfg *config.Config, method, url string, payload interface{}) ([]byte, error) {
	var reqBody io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(cfg.Client.ServerHeaderKeyUserUUID, cfg.Client.ServerHeaderValueUserUUID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
