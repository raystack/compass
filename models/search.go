package models

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
}

// SearchResult represents an individual result item
type SearchResult struct {
	TypeName string
	Record   Record
}

// RecordSearcher is an interface representing the ability
// to search records. The search is intended to be fuzzy over
// the fields of the records, while also supporting filter criteria
type RecordSearcher interface {
	Search(cfg SearchConfig) ([]SearchResult, error)
}
