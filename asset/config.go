package asset

import (
	"github.com/odpf/columbus/validator"
)

type Config struct {
	Types         []Type
	Services      []string
	Size          int
	Offset        int
	SortBy        string `validate:"omitempty,oneof=name type service created_at updated_at"`
	SortDirection string `validate:"omitempty,oneof=asc desc"`
	QueryFields   []string
	Query         string
	Data          map[string]string
}

// GRPCConfig will be refactored to the config above
type GRPCConfig struct {
	Text    string `json:"text"`
	Type    Type   `json:"type"`
	Service string `json:"service"`
	Size    int    `json:"size"`
	Offset  int    `json:"offset"`
}

func (cfg *Config) Validate() error {
	return validator.ValidateStruct(cfg)
}

func (cfg *Config) AssignDefault() {
	if len(cfg.Data) == 0 {
		cfg.Data = nil
	}
}

func (grpcCfg GRPCConfig) ToConfig() Config {
	cfg := Config{
		Size:   grpcCfg.Size,
		Offset: grpcCfg.Offset,
	}

	if len(grpcCfg.Type) > 0 {
		cfg.Types = []Type{grpcCfg.Type}
	}
	if len(grpcCfg.Service) > 0 {
		cfg.Services = []string{grpcCfg.Service}
	}

	if len(grpcCfg.Text) > 0 {
		cfg.Query = grpcCfg.Text
	}

	return cfg
}
