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

func (cfg *Config) Validate() error {
	return validator.ValidateStruct(cfg)
}

func (cfg *Config) AssignDefault() {
	if len(cfg.Data) == 0 {
		cfg.Data = nil
	}
}
