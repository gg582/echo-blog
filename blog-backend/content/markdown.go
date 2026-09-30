// Package content reads and writes the markdown files behind posts and the
// about/contact pages, and renders them into models.Post.
package content

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"

	"github.com/gg582/echo-blog/blog-backend/models"
)

// DefaultAuthor is used when a markdown file has no author front matter.
const DefaultAuthor = "블로그 관리자"

// authorRegex matches the "--- author: <name> ---" front matter at the start of a file.
var authorRegex = regexp.MustCompile(`(?s)^---\s*author:\s*(.+?)\s*---[\r\n]*`)

// SplitAuthor extracts the author from the front matter and returns it with
// the front matter removed from the content.
func SplitAuthor(content []byte) (author string, body []byte) {
	s := string(content)
	matches := authorRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		return DefaultAuthor, content
	}
	return strings.TrimSpace(matches[1]), []byte(authorRegex.ReplaceAllString(s, ""))
}

// Title returns the first non-blank line of body, without a leading "#".
// If body has no such line, fallback is returned.
func Title(body []byte, fallback string) string {
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimPrefix(line, "#"))
		}
		return line
	}
	return fallback
}

// Render turns raw markdown into a Post.
func Render(id, fileName string, raw []byte, modTime time.Time, fallbackTitle string) models.Post {
	author, body := SplitAuthor(raw)
	return models.Post{
		ID:          id,
		Title:       Title(body, fallbackTitle),
		ContentHTML: string(blackfriday.Run(body)),
		Author:      author,
		CreatedAt:   modTime,
		FileName:    fileName,
	}
}

// NewPostMarkdown formats a new post as stored on disk: author front matter,
// the title as a heading, then the content.
func NewPostMarkdown(req models.NewPostRequest) []byte {
	return fmt.Appendf(nil, "---\nauthor: %s\n---\n\n# %s\n\n%s", req.Author, req.Title, req.Content)
}
