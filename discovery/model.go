package discovery

type SearchResult struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Type        string            `json:"type"`
	Service     string            `json:"service"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}
