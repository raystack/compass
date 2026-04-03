package config

import (
	"fmt"

	"github.com/raystack/compass/core/embedding"
	"github.com/raystack/compass/internal/client"
	"github.com/raystack/compass/internal/telemetry"
	"github.com/raystack/compass/store/postgres"
)

// Config is the root configuration for the compass application.
type Config struct {
	LogLevel  string           `yaml:"log_level" mapstructure:"log_level" default:"info"`
	Telemetry telemetry.Config `mapstructure:"telemetry"`
	DB        postgres.Config  `mapstructure:"db"`
	Service   ServerConfig     `mapstructure:"service"`
	Client    client.Config    `mapstructure:"client"`
	Embedding EmbeddingConfig  `mapstructure:"embedding"`
}

// EmbeddingConfig configures the embedding pipeline.
type EmbeddingConfig struct {
	Enabled   bool                 `yaml:"enabled" mapstructure:"enabled" default:"false"`
	Provider  string               `yaml:"provider" mapstructure:"provider" default:"ollama"`
	Ollama    embedding.OllamaConfig `yaml:"ollama" mapstructure:"ollama"`
	OpenAI    embedding.OpenAIConfig `yaml:"openai" mapstructure:"openai"`
	Workers   int                  `yaml:"workers" mapstructure:"workers" default:"2"`
	QueueSize int                  `yaml:"queue_size" mapstructure:"queue_size" default:"1000"`
	MaxTokens int                  `yaml:"max_tokens" mapstructure:"max_tokens" default:"512"`
	Overlap   int                  `yaml:"overlap" mapstructure:"overlap" default:"50"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host    string `mapstructure:"host" default:"0.0.0.0"`
	Port    int    `mapstructure:"port" default:"8080"`
	BaseUrl string `mapstructure:"baseurl" default:"localhost:8080"`

	// User Identity
	Identity IdentityConfig `mapstructure:"identity"`

	// CORS
	CORS CORSConfig `mapstructure:"cors"`

	// Message size limits (for compatibility)
	MaxRecvMsgSize int `yaml:"max_recv_msg_size" mapstructure:"max_recv_msg_size" default:"33554432"`
	MaxSendMsgSize int `yaml:"max_send_msg_size" mapstructure:"max_send_msg_size" default:"33554432"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" mapstructure:"allowed_origins" default:"[*]"`
}

func (cfg ServerConfig) Addr() string { return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port) }

type IdentityConfig struct {
	HeaderKeyUserUUID   string `yaml:"headerkey_uuid" mapstructure:"headerkey_uuid" default:"Compass-User-UUID"`
	HeaderValueUserUUID string `yaml:"headervalue_uuid" mapstructure:"headervalue_uuid" default:"raystack@email.com"`
	HeaderKeyUserEmail  string `yaml:"headerkey_email" mapstructure:"headerkey_email" default:"Compass-User-Email"`
	ProviderDefaultName string `yaml:"provider_default_name" mapstructure:"provider_default_name" default:""`

	NamespaceClaimKey string `yaml:"namespace_claim_key" mapstructure:"namespace_claim_key" default:"namespace_id"`
}
