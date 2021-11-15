package record

type Type string

func (t Type) String() string {
	return string(t)
}

const (
	TypeTable     Type = "table"
	TypeJob       Type = "job"
	TypeDashboard Type = "dashboard"
	TypeTopic     Type = "topic"
)

var TypeList = []Type{
	TypeTable,
	TypeJob,
	TypeDashboard,
	TypeTopic,
}
