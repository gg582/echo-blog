// Package assets manages uploaded files: listing, saving and deleting them in
// the assets directory.
package assets

import (
	"context"
	"errors"
	"io"
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
	// ErrExists is returned when renaming onto a file that already exists.
	ErrExists = errors.New("file already exists")
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
		files = append(files, fileInfo(info))
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt > files[j].ModifiedAt
	})
	return files, nil
}

// Stat returns the metadata of the named file.
func (s *Store) Stat(name string) (models.FileInfo, error) {
	if !validName(name) {
		return models.FileInfo{}, ErrInvalidName
	}
	info, err := os.Stat(filepath.Join(s.dir, name))
	if errors.Is(err, fs.ErrNotExist) {
		return models.FileInfo{}, ErrNotFound
	}
	if err != nil {
		return models.FileInfo{}, err
	}
	return fileInfo(info), nil
}

// Delete removes the named file.
func (s *Store) Delete(name string) error {
	if !validName(name) {
		return ErrInvalidName
	}
	path := filepath.Join(s.dir, name)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return ErrNotFound
	}
	return os.Remove(path)
}

// Replace overwrites the content of an existing file, keeping its name so
// links to it stay valid. The new content is written to a temporary file and
// renamed into place, so readers never see a partial file.
func (s *Store) Replace(name string, content io.Reader) (models.FileInfo, error) {
	if _, err := s.Stat(name); err != nil {
		return models.FileInfo{}, err
	}

	tmp, err := os.CreateTemp(s.dir, ".replace-*")
	if err != nil {
		return models.FileInfo{}, err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed
	if _, err := io.Copy(tmp, content); err != nil {
		tmp.Close()
		return models.FileInfo{}, err
	}
	if err := tmp.Close(); err != nil {
		return models.FileInfo{}, err
	}
	// CreateTemp uses mode 0600; match files saved by uploads.
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return models.FileInfo{}, err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(s.dir, name)); err != nil {
		return models.FileInfo{}, err
	}
	return s.Stat(name)
}

// Rename renames a file. It fails with ErrExists rather than overwrite
// another file.
func (s *Store) Rename(from, to string) (models.FileInfo, error) {
	if !validName(to) {
		return models.FileInfo{}, ErrInvalidName
	}
	if _, err := s.Stat(from); err != nil {
		return models.FileInfo{}, err
	}
	if from == to {
		return s.Stat(to)
	}
	// Link fails if the target exists, so a concurrent upload is never clobbered.
	oldPath, newPath := filepath.Join(s.dir, from), filepath.Join(s.dir, to)
	err := os.Link(oldPath, newPath)
	switch {
	case errors.Is(err, fs.ErrExist):
		return models.FileInfo{}, ErrExists
	case err == nil:
		err = os.Remove(oldPath)
	default:
		// Filesystems without hard links: check, then rename.
		if _, statErr := os.Lstat(newPath); statErr == nil {
			return models.FileInfo{}, ErrExists
		}
		err = os.Rename(oldPath, newPath)
	}
	if err != nil {
		return models.FileInfo{}, err
	}
	return s.Stat(to)
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

// validName reports whether name is a plain file name inside the directory.
func validName(name string) bool {
	return name != "" && !strings.ContainsAny(name, `/\`) && !strings.Contains(name, "..")
}

func fileInfo(info fs.FileInfo) models.FileInfo {
	return models.FileInfo{
		Name:       info.Name(),
		Size:       info.Size(),
		ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
	}
}
