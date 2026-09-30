package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gg582/echo-blog/blog-backend/config"
)

// shutdownTimeout bounds how long in-flight requests may take after ctx ends.
const shutdownTimeout = 10 * time.Second

// Run serves the app according to its config until ctx is cancelled, then
// shuts down gracefully. It returns nil on a clean shutdown.
func (a *App) Run(ctx context.Context) error {
	cfg := a.cfg
	scheme := "HTTP"
	if cfg.UseHTTPS {
		scheme = "HTTPS"
	}
	log.Printf("Server starting on %s (%s)...", cfg.ServerAddr, scheme)

	srv := &http.Server{Addr: cfg.ServerAddr, Handler: a.handler}
	switch {
	case !cfg.UseHTTPS:
		return serve(ctx, srv, srv.ListenAndServe)
	case fileExists(cfg.TLSCertFile) && fileExists(cfg.TLSKeyFile):
		log.Printf("Using local TLS certificate for %s.", cfg.TLSDomain)
		return serve(ctx, srv, func() error { return srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile) })
	default:
		return serveWithAutocert(ctx, cfg, srv)
	}
}

// serve runs listen until it fails or ctx ends, in which case srv is shut
// down gracefully.
func serve(ctx context.Context, srv *http.Server, listen func() error) error {
	errc := make(chan error, 1)
	go func() { errc <- listen() }()

	select {
	case err := <-errc:
		return fmt.Errorf("server failed on %s: %w", srv.Addr, err)
	case <-ctx.Done():
	}

	log.Printf("Shutting down server on %s...", srv.Addr)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down server on %s: %w", srv.Addr, err)
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// serveWithAutocert obtains a certificate from Let's Encrypt via the HTTP-01
// challenge and serves HTTPS with it.
func serveWithAutocert(ctx context.Context, cfg *config.Config, srv *http.Server) error {
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
	challengeSrv := &http.Server{Addr: cfg.HTTPChallengeAddr, Handler: manager.HTTPHandler(nil)}

	// The challenge server lives exactly as long as the HTTPS server.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	challengeErr := make(chan error, 1)
	go func() {
		log.Printf("HTTP-01 challenge server listening on %s.", cfg.HTTPChallengeAddr)
		challengeErr <- serve(ctx, challengeSrv, func() error { return challengeSrv.Serve(challengeListener) })
	}()

	srv.TLSConfig = &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: manager.GetCertificate,
	}
	err = serve(ctx, srv, func() error { return srv.ListenAndServeTLS("", "") })
	cancel()
	if cerr := <-challengeErr; cerr != nil {
		log.Printf("HTTP-01 challenge server: %v", cerr)
	}
	return err
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
