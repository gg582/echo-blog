package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/database"
	"golang.org/x/crypto/acme/autocert"
)

// Serve starts the blog server according to cfg. It returns only when the
// server stops, with a non-nil error unless shutdown was clean.
func Serve(cfg *config.Config, handler http.Handler) error {
	database.InitDatabase(cfg.DBPath)
	log.Println("Database loaded.")

	scheme := "HTTP"
	if cfg.UseHTTPS {
		scheme = "HTTPS"
	}
	log.Printf("Server starting on %s (%s)...", cfg.ServerAddr, scheme)

	if !cfg.UseHTTPS {
		return listenAndServe(cfg.ServerAddr, handler)
	}

	if fileExists(cfg.TLSCertFile) && fileExists(cfg.TLSKeyFile) {
		log.Printf("Using local TLS certificate for %s.", cfg.TLSDomain)
		return http.ListenAndServeTLS(cfg.ServerAddr, cfg.TLSCertFile, cfg.TLSKeyFile, handler)
	}

	return serveWithAutocert(cfg, handler)
}

// serveWithAutocert obtains a certificate from Let's Encrypt via the HTTP-01
// challenge and serves HTTPS with it.
func serveWithAutocert(cfg *config.Config, handler http.Handler) error {
	if err := os.MkdirAll(cfg.ACMECacheDir, 0o700); err != nil {
		return fmt.Errorf("create autocert cache directory %s: %w", cfg.ACMECacheDir, err)
	}

	log.Printf("Requesting Let's Encrypt certificate for %s...", cfg.TLSDomain)
	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.TLSDomain),
		Cache:      autocert.DirCache(cfg.ACMECacheDir),
	}

	challengeListener, err := net.Listen("tcp", cfg.HTTPChallengeAddr)
	if err != nil {
		return fmt.Errorf("bind HTTP-01 challenge server on %s: %w", cfg.HTTPChallengeAddr, err)
	}
	challengeServer := &http.Server{Handler: manager.HTTPHandler(nil)}

	challengeErrChan := make(chan error, 1)
	go func() {
		log.Printf("HTTP-01 challenge server listening on %s.", cfg.HTTPChallengeAddr)
		if serveErr := challengeServer.Serve(challengeListener); serveErr != nil && serveErr != http.ErrServerClosed {
			challengeErrChan <- serveErr
		}
	}()
	select {
	case serveErr := <-challengeErrChan:
		return fmt.Errorf("HTTP-01 challenge server failed: %w", serveErr)
	default:
	}

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: handler,
		TLSConfig: &tls.Config{
			MinVersion:     tls.VersionTLS12,
			GetCertificate: manager.GetCertificate,
		},
	}

	serveErr := server.ListenAndServeTLS("", "")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if shutdownErr := challengeServer.Shutdown(shutdownCtx); shutdownErr != nil && shutdownErr != http.ErrServerClosed {
		log.Printf("Failed to shut down challenge server cleanly: %v", shutdownErr)
	}

	return serveErr
}

func listenAndServe(addr string, handler http.Handler) error {
	if err := http.ListenAndServe(addr, handler); err != nil {
		return fmt.Errorf("server failed on %s: %w", addr, err)
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
