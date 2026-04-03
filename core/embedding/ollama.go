package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaConfig configures the Ollama embedding provider.
type OllamaConfig struct {
	Host  string `yaml:"host" mapstructure:"host" default:"http://localhost:11434"`
	Model string `yaml:"model" mapstructure:"model" default:"nomic-embed-text"`
}

// Ollama generates embeddings using a local Ollama server.
// No API key required.
type Ollama struct {
	cfg    OllamaConfig
	client *http.Client
}

func NewOllama(cfg OllamaConfig) *Ollama {
	if cfg.Host == "" {
		cfg.Host = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "nomic-embed-text"
	}
	return &Ollama{cfg: cfg, client: &http.Client{}}
}

func (o *Ollama) Name() string { return "ollama/" + o.cfg.Model }

func (o *Ollama) Dimensions() int {
	// nomic-embed-text produces 768-dim vectors
	return 768
}

func (o *Ollama) Embed(ctx context.Context, text string) ([]float32, error) {
	vectors, err := o.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("ollama: no embedding returned")
	}
	return vectors[0], nil
}

func (o *Ollama) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(ollamaEmbedRequest{
		Model: o.cfg.Model,
		Input: texts,
	})
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.Host+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama: decode response: %w", err)
	}

	vectors := make([][]float32, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		vec := make([]float32, len(emb))
		for j, v := range emb {
			vec[j] = float32(v)
		}
		vectors[i] = vec
	}
	return vectors, nil
}

type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}
