package server

import (
	"database/sql"
	"net/http"

	"github.com/gg582/echo-blog/blog-backend/assets"
	"github.com/gg582/echo-blog/blog-backend/auth"
	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/content"
	"github.com/gg582/echo-blog/blog-backend/database"
	"github.com/gg582/echo-blog/blog-backend/handlers"
	"github.com/gg582/echo-blog/blog-backend/workerpool"
)

const (
	uploadWorkers   = 5
	uploadQueueSize = 48
)

// App wires the configuration, storage and HTTP handler together.
type App struct {
	cfg     *config.Config
	db      *sql.DB
	pool    *workerpool.Pool
	handler http.Handler
}

// NewApp opens the database, starts the upload workers and builds the router.
// Call Close when done.
func NewApp(cfg *config.Config) (*App, error) {
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	tokens, err := auth.NewSigner(cfg.AuthSecret)
	if err != nil {
		db.Close()
		return nil, err
	}

	pool := workerpool.New(uploadWorkers, uploadQueueSize)
	h := &handlers.Handlers{
		Posts:   content.NewPosts(cfg.PostsDir),
		About:   content.Page{ID: "about", Path: cfg.AboutMD, DefaultTitle: "About Us"},
		Contact: content.Page{ID: "contact", Path: cfg.ContactMD, DefaultTitle: "Contact Us"},
		Assets:  assets.NewStore(cfg.AssetsDir, pool),
		Users:   auth.NewUsers(db),
		Tokens:  tokens,
	}

	return &App{
		cfg:     cfg,
		db:      db,
		pool:    pool,
		handler: NewRouter(cfg, h),
	}, nil
}

// Handler returns the HTTP handler serving the API, assets and frontend.
func (a *App) Handler() http.Handler {
	return a.handler
}

// Close waits for pending uploads and closes the database.
func (a *App) Close() error {
	a.pool.Close()
	return a.db.Close()
}
