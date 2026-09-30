package content

import (
	"path/filepath"

	"github.com/gg582/echo-blog/blog-backend/models"
)

// Page is a standalone markdown page such as about or contact.
type Page struct {
	// ID is the id reported in the rendered Post.
	ID string
	// Path is the markdown file backing the page.
	Path string
	// DefaultTitle is used when the file has no non-blank line.
	DefaultTitle string
}

// Load renders the page. It returns ErrNotFound if the file does not exist.
func (pg Page) Load() (models.Post, error) {
	return renderFile(pg.Path, pg.ID, filepath.Base(pg.Path), pg.DefaultTitle)
}
