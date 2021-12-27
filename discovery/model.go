package discovery

import "github.com/odpf/columbus/record"

// RecordFilter is a filter intended to be used as a search
// criteria for operations involving record search
type RecordFilter = map[string][]string

// SearchResult represents an item/result in a list of search results
type SearchResult struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Type        string            `json:"type"`
	Service     string            `json:"service"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

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
	Queries map[string]string
}

// GetConfig represents a get query along
// with any corresponding filter(s)
type GetConfig struct {
	// Filters specifies document level values to look for.
	// Multiple values can be specified for a single key
	Filters RecordFilter

	// Number of relevant results to return
	Size int

	// Offset to fetch records from
	From int
}

// RecordList is a struct that wraps list of records with total
type RecordList struct {
	// Data contains list of fetched records
	Data []record.Record `json:"data"`

	// Count is the length of Data
	Count int `json:"count"`

	// Total is the total of available data in the repository
	// It also includes those that are not fetched
	Total int `json:"total"`
}
