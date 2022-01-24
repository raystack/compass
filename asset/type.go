package asset

import (
	"fmt"
)

const (
	TypeTable     Type = "table"
	TypeJob       Type = "job"
	TypeDashboard Type = "dashboard"
	TypeTopic     Type = "topic"
)

// AllSupportedTypes holds a list of all supported types struct
var AllSupportedTypes = []Type{
	TypeTable,
	TypeJob,
	TypeDashboard,
	TypeTopic,
}

// Type specifies a supported type name
type Type string

// String cast Type to string
func (t Type) String() string {
	return string(t)
}

// IsValid will validate whether the typename is valid or not
func (t Type) IsValid() error {
	switch t {
	case TypeTable, TypeJob, TypeDashboard, TypeTopic:
		return nil
	}
	return fmt.Errorf("invalid type name: %s", t)
}
