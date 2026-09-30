package server

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// requestLogger logs one line per request in the format the chi logger used,
// so existing server.log readers keep working:
//
//	"GET http://host/path HTTP/1.1" from 1.2.3.4:5678 - 200 123B in 1.2ms
func requestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:     true,
		LogStatus:       true,
		LogResponseSize: true,
		LogLatency:      true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			r := c.Request()
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			log.Printf("\"%s\" from %s - %03d %dB in %s",
				r.Method+" "+scheme+"://"+r.Host+r.RequestURI+" "+r.Proto,
				r.RemoteAddr, v.Status, v.ResponseSize, v.Latency)
			return nil
		},
	})
}

// methodNotAllowed answers 405 with an Allow header listing the methods that
// would be accepted for the path. The SPA route accepts GET on every path, so
// GET is always listed, as chi did.
func methodNotAllowed(c *echo.Context) error {
	var methods []string
	if allow, ok := c.Get(echo.ContextKeyHeaderAllow).(string); ok {
		for m := range strings.SplitSeq(allow, ", ") {
			if m != http.MethodOptions && m != "" {
				methods = append(methods, m)
			}
		}
	}
	if !slices.Contains(methods, http.MethodGet) {
		methods = append(methods, http.MethodGet)
	}
	c.Response().Header().Set(echo.HeaderAllow, strings.Join(methods, ", "))
	return echo.ErrMethodNotAllowed
}
