package embedding

import "context"

// Provider generates vector embeddings from text.
type Provider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
	Name() string
}

// EmbeddingFunc generates vector embeddings from text.
// Used by HybridSearch for query-time embedding.
type EmbeddingFunc func(ctx context.Context, text string) ([]float32, error)

// AsEmbeddingFunc adapts a Provider to an EmbeddingFunc.
func AsEmbeddingFunc(p Provider) EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		return p.Embed(ctx, text)
	}
}
