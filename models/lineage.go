package models

// DataflowDir describes the direction of data IO with regards to another resource
// it's "upstream" if the application only reads, "downstream" if the application only writes
// "bidirectional" if the application both reads and writes to another resource.
// "bidirectional" direction is currently not supported.
type DataflowDir string

func (d DataflowDir) Valid() bool {
	for _, dir := range AllDataflowDir {
		if d == dir {
			return true
		}
	}
	return false
}

const (
	DataflowDirUpstream   = DataflowDir("upstream")
	DataflowDirDownstream = DataflowDir("downstream")
)

var AllDataflowDir = []DataflowDir{DataflowDirDownstream, DataflowDirUpstream}

// LineageDescriptor describes an association
// between one type to another
type LineageDescriptor struct {
	// Type is the name of the type class that
	// is being referred
	Type string `json:"type"`

	// Query is a JSON Path query that is run against
	// a RecordV1 to obtain the id's of it's related resources
	Query string `json:"query"`

	// Direction of the related resource in terms of whether
	// it's upstream or downstream of the current resource
	Dir DataflowDir `json:"direction"`
}
