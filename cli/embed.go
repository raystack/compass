package cli

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/core/chunking"
	"github.com/raystack/compass/core/document"
	"github.com/raystack/compass/core/embedding"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/config"
	compassserver "github.com/raystack/compass/internal/server"
	"github.com/raystack/compass/store/postgres"
	"github.com/spf13/cobra"
)

func embedCommand(cfg *config.Config) *cobra.Command {
	var embedType string
	var batchSize int

	cmd := &cobra.Command{
		Use:   "embed",
		Short: "Generate embeddings for entities and documents",
		Long:  "Batch generate embeddings for all entities and/or documents. Requires an embedding provider configured.",
		Example: heredoc.Doc(`
			$ compass embed
			$ compass embed --type entity
			$ compass embed --type document
			$ compass embed --batch-size 50
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			compassserver.InitLogger(cfg.LogLevel)
			return runEmbed(cmd.Context(), cfg, embedType, batchSize)
		},
	}

	cmd.Flags().StringVar(&embedType, "type", "all", "Type to embed: entity, document, or all")
	cmd.Flags().IntVar(&batchSize, "batch-size", 100, "Number of items to process per batch")

	return cmd
}

func runEmbed(ctx context.Context, cfg *config.Config, embedType string, batchSize int) error {
	if !cfg.Embedding.Enabled {
		return fmt.Errorf("embedding is not enabled in config")
	}

	// Init embedding provider
	provider, err := initProvider(cfg.Embedding)
	if err != nil {
		return err
	}
	slog.Info("embedding provider initialized", "provider", provider.Name())

	// Init postgres
	pgClient, err := postgres.NewClient(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pgClient.Close()

	embeddingRepo, err := postgres.NewEmbeddingRepository(pgClient)
	if err != nil {
		return err
	}

	ns := namespace.DefaultNamespace

	if embedType == "all" || embedType == "entity" {
		entityRepo, err := postgres.NewEntityRepository(pgClient)
		if err != nil {
			return err
		}
		if err := embedEntities(ctx, entityRepo, embeddingRepo, provider, ns, batchSize, cfg.Embedding); err != nil {
			return fmt.Errorf("embed entities: %w", err)
		}
	}

	if embedType == "all" || embedType == "document" {
		docRepo, err := postgres.NewDocumentRepository(pgClient)
		if err != nil {
			return err
		}
		if err := embedDocuments(ctx, docRepo, embeddingRepo, provider, ns, batchSize, cfg.Embedding); err != nil {
			return fmt.Errorf("embed documents: %w", err)
		}
	}

	slog.Info("embedding complete")
	return nil
}

func embedEntities(ctx context.Context, entityRepo entity.Repository, embeddingRepo embedding.Repository,
	provider embedding.Provider, ns *namespace.Namespace, batchSize int, cfg config.EmbeddingConfig) error {

	offset := 0
	total := 0

	for {
		entities, err := entityRepo.GetAll(ctx, ns, entity.Filter{Size: batchSize, Offset: offset})
		if err != nil {
			return err
		}
		if len(entities) == 0 {
			break
		}

		for _, ent := range entities {
			chunks := chunking.SerializeEntity(ent)
			if len(chunks) == 0 {
				continue
			}

			text := chunks[0].Context + "\n\n" + chunks[0].Content
			vec, err := provider.Embed(ctx, text)
			if err != nil {
				slog.Error("failed to embed entity", "urn", ent.URN, "error", err)
				continue
			}

			embs := []embedding.Embedding{{
				EntityURN:   ent.URN,
				ContentID:   ent.ID,
				ContentType: "entity",
				Content:     chunks[0].Content,
				Context:     chunks[0].Context,
				Vector:      vec,
				Position:    0,
				Heading:     chunks[0].Heading,
				TokenCount:  chunking.EstimateTokens(chunks[0].Content),
			}}

			if err := embeddingRepo.UpsertBatch(ctx, ns, embs); err != nil {
				slog.Error("failed to store entity embedding", "urn", ent.URN, "error", err)
				continue
			}
			total++
		}

		slog.Info("embedded entities", "batch", offset/batchSize+1, "count", len(entities), "total", total)
		offset += batchSize

		if len(entities) < batchSize {
			break
		}
	}

	slog.Info("entity embedding complete", "total", total)
	return nil
}

func embedDocuments(ctx context.Context, docRepo document.Repository, embeddingRepo embedding.Repository,
	provider embedding.Provider, ns *namespace.Namespace, batchSize int, cfg config.EmbeddingConfig) error {

	docs, err := docRepo.GetAll(ctx, ns, document.Filter{Size: 10000})
	if err != nil {
		return err
	}

	total := 0
	for _, doc := range docs {
		chunks := chunking.SplitDocument(doc.Title, doc.Body, chunking.Options{
			MaxTokens: cfg.MaxTokens,
			Overlap:   cfg.Overlap,
			Title:     doc.Title,
		})
		if len(chunks) == 0 {
			continue
		}

		// Prepare texts for batch embedding
		texts := make([]string, len(chunks))
		for i, c := range chunks {
			if c.Context != "" {
				texts[i] = c.Context + "\n\n" + c.Content
			} else {
				texts[i] = c.Content
			}
		}

		vectors, err := provider.EmbedBatch(ctx, texts)
		if err != nil {
			slog.Error("failed to embed document", "title", doc.Title, "entity_urn", doc.EntityURN, "error", err)
			continue
		}

		embs := make([]embedding.Embedding, len(chunks))
		for i, c := range chunks {
			var vec []float32
			if i < len(vectors) {
				vec = vectors[i]
			}
			embs[i] = embedding.Embedding{
				EntityURN:   doc.EntityURN,
				ContentID:   doc.ID,
				ContentType: "document",
				Content:     c.Content,
				Context:     c.Context,
				Vector:      vec,
				Position:    c.Position,
				Heading:     c.Heading,
				TokenCount:  chunking.EstimateTokens(c.Content),
			}
		}

		if err := embeddingRepo.UpsertBatch(ctx, ns, embs); err != nil {
			slog.Error("failed to store document embeddings", "title", doc.Title, "error", err)
			continue
		}
		total += len(embs)
	}

	slog.Info("document embedding complete", "total_chunks", total, "documents", len(docs))
	return nil
}

func initProvider(cfg config.EmbeddingConfig) (embedding.Provider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "openai":
		if cfg.OpenAI.APIKey == "" {
			return nil, fmt.Errorf("openai api_key is required")
		}
		return embedding.NewOpenAI(cfg.OpenAI), nil
	case "ollama", "":
		return embedding.NewOllama(cfg.Ollama), nil
	default:
		return nil, fmt.Errorf("unknown embedding provider: %s", cfg.Provider)
	}
}
