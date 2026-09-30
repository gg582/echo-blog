// Package config loads runtime configuration from environment variables,
// preserving the historical hardcoded behavior when variables are unset.
package config

import (
	"os"
	"strings"
)

// Config holds all runtime configuration for the blog backend.
type Config struct {
	// ServerAddr is the listen address for the HTTP(S) server.
	ServerAddr string
	// DBPath is the path to the SQLite database file.
	DBPath string
	// StaticDir is the directory containing the built frontend assets.
	StaticDir string
	// PostsDir is the directory containing blog post markdown files.
	PostsDir string
	// AssetsDir is the directory where uploaded files are stored.
	AssetsDir string
	// AboutMD is the markdown file backing the about page.
	AboutMD string
	// ContactMD is the markdown file backing the contact page.
	ContactMD string
	// SettingsPath is the JSON file holding site settings edited from the dashboard.
	SettingsPath string
	// UseHTTPS enables TLS serving (local cert files or ACME).
	UseHTTPS bool
	// TLSCertFile is the path to the TLS certificate chain.
	TLSCertFile string
	// TLSKeyFile is the path to the TLS private key.
	TLSKeyFile string
	// TLSDomain is the domain requested via ACME and whitelisted by autocert.
	TLSDomain string
	// ACMECacheDir is the local cache directory for ACME certificates.
	ACMECacheDir string
	// HTTPChallengeAddr is the listen address of the ACME HTTP-01 challenge server.
	HTTPChallengeAddr string
	// AllowedOrigins is the list of CORS-allowed origins. Empty means no CORS middleware.
	AllowedOrigins []string
	// AuthSecret is the HMAC key used to sign auth tokens. Empty generates a random per-process secret.
	AuthSecret string
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load reads configuration from environment variables and returns a Config
// with defaults matching the previous hardcoded values.
func Load() *Config {
	cfg := &Config{
		ServerAddr:        getenv("SERVER_ADDR", ":8080"),
		DBPath:            getenv("DB_PATH", "./auth.db"),
		StaticDir:         getenv("STATIC_DIR", "../blog-frontend/build"),
		PostsDir:          getenv("POSTS_DIR", "./posts"),
		AssetsDir:         getenv("ASSETS_DIR", "./posts/assets"),
		AboutMD:           getenv("ABOUT_MD", "./about/about.md"),
		ContactMD:         getenv("CONTACT_MD", "./contact/contact.md"),
		SettingsPath:      getenv("SETTINGS_PATH", "./settings.json"),
		TLSCertFile:       getenv("TLS_CERT_FILE", "/etc/letsencrypt/live/chatter.pw/fullchain.pem"),
		TLSKeyFile:        getenv("TLS_KEY_FILE", "/etc/letsencrypt/live/chatter.pw/privkey.pem"),
		TLSDomain:         getenv("TLS_DOMAIN", "chatter.pw"),
		ACMECacheDir:      getenv("ACME_CACHE_DIR", "./cert-cache"),
		HTTPChallengeAddr: getenv("HTTP_CHALLENGE_ADDR", ":80"),
	}

	cfg.UseHTTPS = strings.EqualFold(os.Getenv("USE_HTTPS"), "true")
	cfg.AuthSecret = os.Getenv("AUTH_SECRET")

	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, origin)
			}
		}
	}

	return cfg
}
