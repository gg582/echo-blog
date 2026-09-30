package content

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gg582/echo-blog/blog-backend/models"
)

// ErrNotFound is returned when a post or page file does not exist.
var ErrNotFound = errors.New("not found")

// Posts stores posts as <id>.md files in a directory.
type Posts struct {
	dir string
}

// NewPosts returns a Posts store rooted at dir.
func NewPosts(dir string) *Posts {
	return &Posts{dir: dir}
}

func (p *Posts) path(id string) string {
	return filepath.Join(p.dir, id+".md")
}

// List renders every .md file in the directory. Files that cannot be read
// are logged and skipped. It stops early if ctx is cancelled.
func (p *Posts) List(ctx context.Context) ([]models.Post, error) {
	ids, err := p.ids()
	if err != nil {
		return nil, err
	}

	var posts []models.Post
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		post, err := p.Get(id)
		if err != nil {
			log.Printf("error reading file: %s - %v", p.path(id), err)
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// Source is the id and raw markdown of one post.
type Source struct {
	ID       string
	Markdown []byte
}

// Sources returns the raw markdown of every post. Files that cannot be read
// are logged and skipped. It stops early if ctx is cancelled.
func (p *Posts) Sources(ctx context.Context) ([]Source, error) {
	ids, err := p.ids()
	if err != nil {
		return nil, err
	}
	sources := make([]Source, 0, len(ids))
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		raw, err := p.Raw(id)
		if err != nil {
			log.Printf("error reading file: %s - %v", p.path(id), err)
			continue
		}
		sources = append(sources, Source{ID: id, Markdown: raw})
	}
	return sources, nil
}

// ids returns the ids of all posts in directory order.
func (p *Posts) ids() ([]string, error) {
	entries, err := os.ReadDir(p.dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory '%s': %w", p.dir, err)
	}
	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			ids = append(ids, strings.TrimSuffix(entry.Name(), ".md"))
		}
	}
	return ids, nil
}

// Get renders the post with the given id.
func (p *Posts) Get(id string) (models.Post, error) {
	return renderFile(p.path(id), id, id+".md", id)
}

// Raw returns the markdown source of the post with the given id.
func (p *Posts) Raw(id string) ([]byte, error) {
	data, err := os.ReadFile(p.path(id))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

// EnsureDir creates the posts directory if it does not exist.
func (p *Posts) EnsureDir() error {
	return os.MkdirAll(p.dir, os.ModePerm)
}

// Write creates or overwrites the post with the given id.
func (p *Posts) Write(id string, markdown []byte) error {
	return os.WriteFile(p.path(id), markdown, 0o644)
}

// Delete removes the post with the given id.
func (p *Posts) Delete(id string) error {
	path := p.path(id)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return ErrNotFound
	}
	return os.Remove(path)
}

// renderFile reads a markdown file and renders it into a Post.
func renderFile(path, id, fileName, fallbackTitle string) (models.Post, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return models.Post{}, ErrNotFound
	}
	if err != nil {
		return models.Post{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return models.Post{}, err
	}
	return Render(id, fileName, raw, info.ModTime(), fallbackTitle), nil
}
