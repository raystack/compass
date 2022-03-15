package asset

const (
	TypeTable     Type = "table"
	TypeJob       Type = "job"
	TypeDashboard Type = "dashboard"
	TypeTopic     Type = "topic"
)

// AllSupportedTypes holds a list of all supported types struct
var AllSupportedTypes = []Type{
	TypeTable.String(),
	TypeJob.String(),
	TypeDashboard.String(),
	TypeTopic.String(),
}

// Type specifies a supported type name
type Type string

// String cast Type to string
func (t Type) String() string {
	return string(t)
}

// IsValid will validate whether the typename is valid or not
func (t Type) IsValid() bool {
	switch t {
	case TypeTable, TypeJob, TypeDashboard, TypeTopic:
		return true
	}
	return false
}

func GetTypeEnum(t string) Type {
	switch {
	case t == TypeTable.String():
		return TypeTable
	case t == TypeJob.String():
		return TypeJob
	case t == TypeDashboard.String():
		return TypeDashboard
	case t == TypeTopic.String():
		return TypeTopic
	}
}
