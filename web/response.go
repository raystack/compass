package web

// ErrorResponse defines the JSON message returned
// from handlers when an error occurs
type ErrorResponse struct {
	Reason string `json:"reason"`
}

// StatusResponse is a generic message for reporting
// the status of an operation
type StatusResponse struct {
	Status string `json:"status"`
}

// ValidationErrorResponse defines
type ValidationErrorResponse struct {
	ErrorResponse

	// details represent(s) the failed records.
	// the key is the index of the record, and the value
	// is an error message explaining the problem
	Details map[int]string `json:"details"`
}

func NewValidationErrorResponse(details map[int]string) *ValidationErrorResponse {
	res := &ValidationErrorResponse{}
	res.Reason = "validation error"
	res.Details = details
	return res
}

// SearchResponse defines an individual item
// in the search response
type SearchResponse struct {
	Title          string            `json:"title"`
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	Classification string            `json:"classification"`
	Description    string            `json:"description"`
	Labels         map[string]string `json:"labels"`
}
