package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/gg582/echo-blog/blog-backend/auth"
)

// userKey is the context key under which RequireAuth stores the username.
const userKey = "user"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login checks the credentials and returns a signed token.
func (h *Handlers) Login(c *echo.Context) error {
	var req loginRequest
	if err := decodeJSON(c, &req); err != nil {
		return textError(http.StatusBadRequest, "Invalid request body")
	}

	err := h.Users.Authenticate(c.Request().Context(), req.Username, req.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return textError(http.StatusUnauthorized, "Invalid credentials")
	}
	if err != nil {
		log.Printf("login: %v", err)
		return textError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Login succeed",
		"token":   h.Tokens.Issue(req.Username),
	})
}

// RequireAuth only lets requests with a valid "Authorization: Bearer <token>"
// header through, and stores the token's username in the context.
func (h *Handlers) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		token, found := strings.CutPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
		if !found || token == "" {
			return textError(http.StatusUnauthorized, "Unauthorized")
		}
		user, ok := h.Tokens.Verify(token)
		if !ok {
			return textError(http.StatusUnauthorized, "Unauthorized")
		}
		c.Set(userKey, user)
		return next(c)
	}
}

// currentUser returns the username set by RequireAuth.
func currentUser(c *echo.Context) string {
	user, _ := c.Get(userKey).(string)
	return user
}
