package server

import (
	"net/http"
	"path/filepath"

	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/handlers"
	"github.com/gg582/echo-blog/blog-backend/utils"
	"github.com/gg582/echo-blog/blog-backend/workerpool"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

const (
	numWorkers   = 5
	jobQueueSize = 48
)

func NewRouter(cfg *config.Config) http.Handler {
	handlers.FileJobQueue = make(chan workerpool.UploadJob, jobQueueSize)
	workerpool.NewWorkerPool(numWorkers, handlers.FileJobQueue)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if len(cfg.AllowedOrigins) > 0 {
		r.Use(cors.New(cors.Options{
			AllowedOrigins:   cfg.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
			AllowCredentials: true,
			MaxAge:           3600,
		}).Handler)
	}

	h := handlers.NewHandlers(cfg.PostsDir, cfg.AssetsDir, cfg.AboutMD, cfg.ContactMD)
	utils.SetAuthSecret(cfg.AuthSecret)

	r.Post("/api/posts", h.GetPostsHandler)
	r.Post("/api/posts/{id}", h.GetPostByIDHandler)
	r.Get("/api/posts/{id}/raw", h.GetRawPostHandler)
	r.Post("/api/edit-post/{id}", handlers.RequireAuth(h.EditPostHandler))
	r.Post("/api/delete-post/{id}", handlers.RequireAuth(h.DeletePostHandler))
	r.Get("/api/files", handlers.RequireAuth(h.GetFilesHandler))
	r.Post("/api/delete-file", handlers.RequireAuth(h.DeleteFileHandler))
	r.Get("/api/about", h.GetAboutPageHandler)
	r.Get("/api/contact", h.GetContactPageHandler)
	r.Post("/api/new-post/{id}", h.CreateNewPostHandler)
	r.Post("/api/upload-file", h.UploadFile)
	r.Post("/api/login", handlers.LoginHandler)

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsDir))))
	r.Get("/*", spaFileHandler(cfg.StaticDir))

	return r
}

// spaFileHandler serves files from staticDir and falls back to index.html
// for paths that match no file, so client-side routes resolve in the SPA.
func spaFileHandler(staticDir string) http.HandlerFunc {
	staticFS := http.FileServer(http.Dir(staticDir))
	indexHTML := filepath.Join(staticDir, "index.html")

	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.ServeFile(w, req, indexHTML)
			return
		}

		f, err := http.Dir(staticDir).Open(req.URL.Path)
		if err != nil {
			http.ServeFile(w, req, indexHTML)
			return
		}
		defer f.Close()

		if stat, err := f.Stat(); err != nil || stat.IsDir() {
			http.ServeFile(w, req, indexHTML)
			return
		}

		staticFS.ServeHTTP(w, req)
	}
}
