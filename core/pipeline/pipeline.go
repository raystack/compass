package pipeline

import (
	"context"
	"log/slog"
	"sync"

	"github.com/raystack/compass/core/chunking"
	"github.com/raystack/compass/core/document"
	"github.com/raystack/compass/core/embedding"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

// Pipeline orchestrates async chunking + embedding on entity/document upsert.
type Pipeline struct {
	repo      embedding.Repository
	provider  embedding.Provider
	queue     chan job
	workers   int
	queueSize int
	maxTokens int
	overlap   int
	wg        sync.WaitGroup
	cancel    context.CancelFunc
}

type job struct {
	Namespace   *namespace.Namespace
	EntityURN   string
	ContentID   string
	ContentType string // "entity" or "document"
	Text        string
	Title       string
}

// Option configures the pipeline.
type Option func(*Pipeline)

func WithWorkers(n int) Option {
	return func(p *Pipeline) {
		if n > 0 {
			p.workers = n
		}
	}
}

func WithQueueSize(n int) Option {
	return func(p *Pipeline) {
		if n > 0 {
			p.queueSize = n
		}
	}
}

func WithMaxTokens(n int) Option {
	return func(p *Pipeline) {
		if n > 0 {
			p.maxTokens = n
		}
	}
}

func WithOverlap(n int) Option {
	return func(p *Pipeline) {
		if n >= 0 {
			p.overlap = n
		}
	}
}

// New creates a new embedding pipeline.
func New(repo embedding.Repository, provider embedding.Provider, opts ...Option) *Pipeline {
	p := &Pipeline{
		repo:      repo,
		provider:  provider,
		workers:   2,
		queueSize: 1000,
		maxTokens: 512,
		overlap:   50,
	}
	for _, opt := range opts {
		opt(p)
	}
	p.queue = make(chan job, p.queueSize)
	return p
}

// Start spawns worker goroutines.
func (p *Pipeline) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
	slog.Info("embedding pipeline started", "workers", p.workers, "queue_size", p.queueSize)
}

// Stop signals workers to finish and waits for completion.
func (p *Pipeline) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	close(p.queue)
	p.wg.Wait()
	slog.Info("embedding pipeline stopped")
}

// EnqueueEntity sends an entity for async embedding.
func (p *Pipeline) EnqueueEntity(ctx context.Context, ns *namespace.Namespace, ent *entity.Entity) error {
	chunks := chunking.SerializeEntity(*ent)
	if len(chunks) == 0 {
		return nil
	}

	// Combine all serialized text
	text := chunks[0].Content

	select {
	case p.queue <- job{
		Namespace:   ns,
		EntityURN:   ent.URN,
		ContentID:   ent.ID,
		ContentType: "entity",
		Text:        text,
		Title:       ent.Name,
	}:
		return nil
	default:
		slog.Warn("embedding pipeline queue full, dropping entity", "urn", ent.URN)
		return nil
	}
}

// EnqueueDocument sends a document for async embedding.
func (p *Pipeline) EnqueueDocument(ctx context.Context, ns *namespace.Namespace, doc *document.Document) error {
	select {
	case p.queue <- job{
		Namespace:   ns,
		EntityURN:   doc.EntityURN,
		ContentID:   doc.ID,
		ContentType: "document",
		Text:        doc.Body,
		Title:       doc.Title,
	}:
		return nil
	default:
		slog.Warn("embedding pipeline queue full, dropping document", "entity_urn", doc.EntityURN, "title", doc.Title)
		return nil
	}
}

func (p *Pipeline) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-p.queue:
			if !ok {
				return
			}
			if err := p.process(ctx, j); err != nil {
				slog.Error("embedding pipeline error",
					"entity_urn", j.EntityURN,
					"content_type", j.ContentType,
					"error", err)
			}
		}
	}
}

func (p *Pipeline) process(ctx context.Context, j job) error {
	var chunks []chunking.Chunk

	switch j.ContentType {
	case "document":
		chunks = chunking.SplitDocument(j.Title, j.Text, chunking.Options{
			MaxTokens: p.maxTokens,
			Overlap:   p.overlap,
			Title:     j.Title,
		})
	default: // "entity"
		// Entity text is already serialized, treat as single chunk
		chunks = []chunking.Chunk{{
			Content:  j.Text,
			Context:  j.Title,
			Heading:  j.Title,
			Position: 0,
		}}
	}

	if len(chunks) == 0 {
		return nil
	}

	// Generate embeddings in batch
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		// Prepend context for better embedding quality
		if c.Context != "" {
			texts[i] = c.Context + "\n\n" + c.Content
		} else {
			texts[i] = c.Content
		}
	}

	vectors, err := p.provider.EmbedBatch(ctx, texts)
	if err != nil {
		return err
	}

	// Build embedding records
	embeddings := make([]embedding.Embedding, len(chunks))
	for i, c := range chunks {
		var vec []float32
		if i < len(vectors) {
			vec = vectors[i]
		}
		embeddings[i] = embedding.Embedding{
			EntityURN:   j.EntityURN,
			ContentID:   j.ContentID,
			ContentType: j.ContentType,
			Content:     c.Content,
			Context:     c.Context,
			Vector:      vec,
			Position:    c.Position,
			Heading:     c.Heading,
			TokenCount:  chunking.EstimateTokens(c.Content),
		}
	}

	return p.repo.UpsertBatch(ctx, j.Namespace, embeddings)
}
