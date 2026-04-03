package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAIConfig configures the OpenAI embedding provider.
type OpenAIConfig struct {
	APIKey  string `yaml:"api_key" mapstructure:"api_key"`
	Model   string `yaml:"model" mapstructure:"model" default:"text-embedding-3-small"`
	BaseURL string `yaml:"base_url" mapstructure:"base_url" default:"https://api.openai.com"`
}

// OpenAI generates embeddings using the OpenAI API.
type OpenAI struct {
	cfg    OpenAIConfig
	client *http.Client
}

func NewOpenAI(cfg OpenAIConfig) *OpenAI {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com"
	}
	if cfg.Model == "" {
		cfg.Model = "text-embedding-3-small"
	}
	return &OpenAI{cfg: cfg, client: &http.Client{}}
}

func (o *OpenAI) Name() string { return "openai/" + o.cfg.Model }

func (o *OpenAI) Dimensions() int {
	// text-embedding-3-small default is 1536, but we request 768 for consistency
	return 768
}

func (o *OpenAI) Embed(ctx context.Context, text string) ([]float32, error) {
	vectors, err := o.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("openai: no embedding returned")
	}
	return vectors[0], nil
}

func (o *OpenAI) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(openaiEmbeddingsRequest{
		Model:      o.cfg.Model,
		Input:      texts,
		Dimensions: o.Dimensions(),
	})
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.BaseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openaiEmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}

	vectors := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		vec := make([]float32, len(d.Embedding))
		for j, v := range d.Embedding {
			vec[j] = float32(v)
		}
		vectors[i] = vec
	}
	return vectors, nil
}

type openaiEmbeddingsRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type openaiEmbeddingsResponse struct {
	Data []openaiEmbeddingData `json:"data"`
}

type openaiEmbeddingData struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}
