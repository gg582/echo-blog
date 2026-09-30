package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/gg582/echo-blog/blog-backend/models" // Import models package
)

// authorRegex helps to find and extract the author from the front matter.
// It looks for "--- author: [username] ---" at the beginning of the file.
var authorRegex = regexp.MustCompile(`(?s)^---\s*author:\s*(.+?)\s*---[\r\n]*`)

// parseAuthorAndCleanContent extracts the author from the content and returns
// the extracted author and the content with the author block removed.
func ParseAuthorAndCleanContent(content []byte) (author string, cleanedContent []byte) {
	contentStr := string(content)
	matches := authorRegex.FindStringSubmatch(contentStr)

	if len(matches) > 1 {
		author = strings.TrimSpace(matches[1]) // Extracted username
		// Remove the matched block from the content
		cleanedContent = []byte(authorRegex.ReplaceAllString(contentStr, ""))
	} else {
		author = "블로그 관리자" // Default author if no block is found
		cleanedContent = content // No block to remove, use original content
	}
	return author, cleanedContent
}

// GetPosts reads and parses all markdown files from the specified directory.
func GetPosts(postsDir string) ([]models.Post, error) {
	var posts []models.Post

	files, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory '%s': %w", postsDir, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue // Skip if it's a directory or not a .md file.
		}

		filePath := filepath.Join(postsDir, file.Name())
		content, err := os.ReadFile(filePath) // Use os.ReadFile
		if err != nil {
			log.Printf("error reading file: %s - %v", filePath, err)
			continue
		}

		author, cleanedContent := ParseAuthorAndCleanContent(content)
		htmlContent := blackfriday.Run(cleanedContent)

		id := strings.TrimSuffix(file.Name(), ".md")
		
		title := strings.TrimSuffix(file.Name(), ".md") // Default title to filename
		lines := strings.Split(string(cleanedContent), "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				if strings.HasPrefix(trimmedLine, "#") {
					title = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "#"))
				} else {
					title = trimmedLine
				}
				break
			}
		}

		fileInfo, _ := os.Stat(filePath) // Get file information for modification time
		createdAt := fileInfo.ModTime()

		posts = append(posts, models.Post{
			ID:          id,
			Title:       title,
			ContentHTML: string(htmlContent),
			Author:      author,
			CreatedAt:   createdAt,
			FileName:    file.Name(),
		})
	}
	return posts, nil
}
