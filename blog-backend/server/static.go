package server

import (
	"net/http"
	"path/filepath"
)

// spaHandler serves files from staticDir and falls back to index.html for
// paths that match no file, so client-side routes resolve in the SPA.
func spaHandler(staticDir string) http.HandlerFunc {
	root := http.Dir(staticDir)
	staticFS := http.FileServer(root)
	indexHTML := filepath.Join(staticDir, "index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || !isFile(root, r.URL.Path) {
			http.ServeFile(w, r, indexHTML)
			return
		}
		staticFS.ServeHTTP(w, r)
	}
}

// isFile reports whether name exists in root and is not a directory.
func isFile(root http.FileSystem, name string) bool {
	f, err := root.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	stat, err := f.Stat()
	return err == nil && !stat.IsDir()
}
