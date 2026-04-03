package chunking

import (
	"fmt"
	"strings"
)

// SplitDocument splits a document body into chunks using markdown structure.
// 1. Split on headings
// 2. If a section exceeds maxTokens, split on paragraphs
// 3. Add contextual prefix to each chunk
func SplitDocument(title, body string, opts Options) []Chunk {
	maxTokens := opts.maxTokens()

	sections := splitOnHeadings(body)
	var chunks []Chunk
	pos := 0

	for _, sec := range sections {
		tokens := EstimateTokens(sec.content)
		prefix := contextPrefix(title, sec.heading)

		if tokens <= maxTokens {
			chunks = append(chunks, Chunk{
				Content:  sec.content,
				Context:  prefix,
				Heading:  sec.heading,
				Position: pos,
			})
			pos++
			continue
		}

		// Section too large — split on paragraphs
		paragraphs := splitOnParagraphs(sec.content)
		for _, para := range mergeParagraphs(paragraphs, maxTokens, opts.overlap()) {
			chunks = append(chunks, Chunk{
				Content:  para,
				Context:  prefix,
				Heading:  sec.heading,
				Position: pos,
			})
			pos++
		}
	}

	return chunks
}

type section struct {
	heading string
	content string
}

// splitOnHeadings splits markdown text into sections bounded by headings.
func splitOnHeadings(text string) []section {
	lines := strings.Split(text, "\n")
	var sections []section
	current := section{heading: "Introduction"}
	var content strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isHeading(trimmed) {
			// Save current section
			if content.Len() > 0 {
				current.content = strings.TrimSpace(content.String())
				if current.content != "" {
					sections = append(sections, current)
				}
			}
			current = section{heading: headingText(trimmed)}
			content.Reset()
		} else {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Save last section
	current.content = strings.TrimSpace(content.String())
	if current.content != "" {
		sections = append(sections, current)
	}

	return sections
}

func isHeading(line string) bool {
	return strings.HasPrefix(line, "# ") ||
		strings.HasPrefix(line, "## ") ||
		strings.HasPrefix(line, "### ") ||
		strings.HasPrefix(line, "#### ")
}

func headingText(line string) string {
	return strings.TrimSpace(strings.TrimLeft(line, "#"))
}

// splitOnParagraphs splits text into paragraphs (separated by blank lines).
func splitOnParagraphs(text string) []string {
	raw := strings.Split(text, "\n\n")
	var paragraphs []string
	for _, p := range raw {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			paragraphs = append(paragraphs, trimmed)
		}
	}
	return paragraphs
}

// mergeParagraphs groups paragraphs into chunks that fit within maxTokens,
// with overlap between adjacent chunks.
func mergeParagraphs(paragraphs []string, maxTokens, overlap int) []string {
	if len(paragraphs) == 0 {
		return nil
	}

	var chunks []string
	var current strings.Builder
	currentTokens := 0

	for i, para := range paragraphs {
		paraTokens := EstimateTokens(para)

		// If a single paragraph exceeds maxTokens, include it as-is
		if paraTokens > maxTokens && current.Len() == 0 {
			chunks = append(chunks, para)
			continue
		}

		if currentTokens+paraTokens > maxTokens && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
			currentTokens = 0

			// Add overlap: include previous paragraph if it fits
			if overlap > 0 && i > 0 {
				prevTokens := EstimateTokens(paragraphs[i-1])
				if prevTokens <= overlap {
					current.WriteString(paragraphs[i-1])
					current.WriteString("\n\n")
					currentTokens = prevTokens
				}
			}
		}

		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(para)
		currentTokens += paraTokens
	}

	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	return chunks
}

func contextPrefix(title, heading string) string {
	if heading == "" || heading == "Introduction" {
		return fmt.Sprintf("Document: %s", title)
	}
	return fmt.Sprintf("Document: %s > Section: %s", title, heading)
}
