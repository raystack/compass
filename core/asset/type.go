package asset

const (
	TypeTable        Type = "table"
	TypeJob          Type = "job"
	TypeDashboard    Type = "dashboard"
	TypeTopic        Type = "topic"
	TypeFeatureTable Type = "feature_table"
	TypeApplication  Type = "application"
	TypeModel        Type = "model"
	TypeQuery        Type = "query"
	TypeMetric       Type = "metric"
	TypeExperiment   Type = "experiment"
)

// AllSupportedTypes holds a list of all supported types struct
var AllSupportedTypes = []Type{
	TypeTable,
	TypeJob,
	TypeDashboard,
	TypeTopic,
	TypeFeatureTable,
	TypeApplication,
	TypeModel,
	TypeQuery,
	TypeMetric,
	TypeExperiment,
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
	case TypeTable, TypeJob, TypeDashboard, TypeTopic,
		TypeFeatureTable, TypeApplication, TypeModel,
		TypeQuery, TypeMetric, TypeExperiment:
		return true
	}
	return false
}
