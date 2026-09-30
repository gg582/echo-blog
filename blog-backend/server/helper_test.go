package server_test

import (
	"net/http"
	"testing"

	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/database"
	"github.com/gg582/echo-blog/blog-backend/server"
	"github.com/gg582/echo-blog/blog-backend/utils"
)

func newTestHandler(t *testing.T, cfg *config.Config, user, password string) http.Handler {
	t.Helper()
	database.InitDatabase(cfg.DBPath)
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec("INSERT INTO blog_users (username, password_hash) VALUES (?, ?)", user, hash); err != nil {
		t.Fatal(err)
	}
	return server.NewRouter(cfg)
}
