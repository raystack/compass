package chunking

// Chunk is a text fragment ready for embedding.
type Chunk struct {
	Content  string // the text to embed
	Context  string // contextual prefix for embedding quality
	Heading  string // section heading
	Position int    // ordering within parent
}

// Options configures chunking behavior.
type Options struct {
	MaxTokens int    // target chunk size in tokens (default 512)
	Overlap   int    // overlap tokens between adjacent chunks (default 50)
	Title     string // parent title for contextual prefix
}

func (o Options) maxTokens() int {
	if o.MaxTokens <= 0 {
		return 512
	}
	return o.MaxTokens
}

func (o Options) overlap() int {
	if o.Overlap <= 0 {
		return 50
	}
	return o.Overlap
}

// EstimateTokens estimates the token count for a text string.
// Uses a simple heuristic: ~1.3 tokens per word on average.
func EstimateTokens(text string) int {
	words := 0
	inWord := false
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			inWord = false
		} else if !inWord {
			inWord = true
			words++
		}
	}
	return (words * 4) / 3
}
