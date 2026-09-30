package server_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gg582/echo-blog/blog-backend/auth"
	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/database"
	"github.com/gg582/echo-blog/blog-backend/server"
)

func newTestHandler(t *testing.T, cfg *config.Config, user, password string) http.Handler {
	t.Helper()
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := auth.NewUsers(db).Create(context.Background(), user, password); err != nil {
		t.Fatal(err)
	}

	app, err := server.NewApp(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close() })
	return app.Handler()
}
