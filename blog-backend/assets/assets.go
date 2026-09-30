// Package assets manages uploaded files: listing, saving and deleting them in
// the assets directory.
package assets

import (
	"context"
	"errors"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gg582/echo-blog/blog-backend/models"
	"github.com/gg582/echo-blog/blog-backend/workerpool"
)

var (
	// ErrNotFound is returned when the named file does not exist.
	ErrNotFound = errors.New("file not found")
	// ErrInvalidName is returned for empty names or names that could escape the directory.
	ErrInvalidName = errors.New("invalid filename")
)

// Store is the assets directory.
type Store struct {
	dir  string
	pool *workerpool.Pool
}

// NewStore returns a Store rooted at dir that saves uploads on pool.
func NewStore(dir string, pool *workerpool.Pool) *Store {
	return &Store{dir: dir, pool: pool}
}

// Dir returns the assets directory.
func (s *Store) Dir() string {
	return s.dir
}

// List returns the files in the directory, newest first. A missing directory
// yields an empty list.
func (s *Store) List() ([]models.FileInfo, error) {
	entries, err := os.ReadDir(s.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []models.FileInfo{}, nil
	}
	if err != nil {
		return nil, err
	}

	files := []models.FileInfo{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, models.FileInfo{
			Name:       info.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt > files[j].ModifiedAt
	})
	return files, nil
}

// Delete removes the named file.
func (s *Store) Delete(name string) error {
	if name == "" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return ErrInvalidName
	}
	path := filepath.Join(s.dir, name)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return ErrNotFound
	}
	return os.Remove(path)
}

// EnsureDir creates the assets directory if it does not exist.
func (s *Store) EnsureDir() error {
	return os.MkdirAll(s.dir, os.ModePerm)
}

// Save stores one uploaded file through the worker pool and returns the name
// it was saved under.
func (s *Store) Save(ctx context.Context, fh *multipart.FileHeader) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	res, err := s.pool.Submit(ctx, workerpool.UploadJob{File: file, FileName: fh.Filename, Dir: s.dir})
	if err != nil {
		return "", err
	}
	return res.SavedFileName, res.Error
}
