package discovery

// RecordFilter is a filter intended to be used as a search
// criteria for operations involving record search
type RecordFilter = map[string][]string

// RecordQuery is a param intended to be used as a match search
// criteria for operations involving record search
type RecordQuery = map[string]string

// SearchConfig represents a search query along
// with any corresponding filter(s)
type SearchConfig struct {
	// Text to search for
	Text string

	// Filters specifies document level values to look for.
	// Multiple values can be specified for a single key
	Filters RecordFilter

	// Number of relevant results to return
	MaxResults int

	// List of record types to search for
	// a zero value signifies that all types should be searched
	TypeWhiteList []string

	// RankBy is a param to rank based on a specific parameter
	RankBy string

	// Queries is a param to search a resource based on record's fields
	Queries RecordQuery
}
