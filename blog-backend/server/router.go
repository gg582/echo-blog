package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/cors"

	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/handlers"
)

// NewRouter builds the Echo instance serving the API, uploaded assets and the
// frontend.
func NewRouter(cfg *config.Config, h *handlers.Handlers) *echo.Echo {
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: errorHandler,
		Router: echo.NewRouter(echo.RouterConfig{
			// Like net/http muxes: OPTIONS without CORS is a 405, not an automatic 204.
			OptionsMethodHandler:    methodNotAllowed,
			MethodNotAllowedHandler: methodNotAllowed,
		}),
	})

	if len(cfg.AllowedOrigins) > 0 {
		// rs/cors runs before routing so preflight requests never reach the router.
		e.Pre(echo.WrapMiddleware(cors.New(cors.Options{
			AllowedOrigins:   cfg.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
			AllowCredentials: true,
			MaxAge:           3600,
		}).Handler))
	}
	e.Use(requestLogger(), middleware.Recover())

	h.Register(e)
	e.Any("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsDir)))))
	e.GET("/*", echo.WrapHandler(spaHandler(cfg.StaticDir)))

	return e
}

// errorHandler renders errors the way net/http does: HTTP errors as a
// text/plain body, 404 as "404 page not found" and 405 with an empty body.
func errorHandler(c *echo.Context, err error) {
	if resp, _ := echo.UnwrapResponse(c.Response()); resp != nil && resp.Committed {
		return
	}
	w, r := c.Response(), c.Request()

	var he *echo.HTTPError
	switch {
	case errors.As(err, &he):
		http.Error(w, he.Message, he.Code)
	case echo.StatusCode(err) == http.StatusNotFound:
		http.NotFound(w, r)
	case echo.StatusCode(err) == http.StatusMethodNotAllowed:
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		log.Printf("%s %s: %v", r.Method, r.URL.RequestURI(), err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
