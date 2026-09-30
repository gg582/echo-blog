package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gg582/echo-blog/blog-backend/database"
	"github.com/gg582/echo-blog/blog-backend/utils"
)


type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    var storedPwHash string
    err := database.DB.QueryRow("SELECT password_hash FROM blog_users WHERE username = ?", req.Username).Scan(&storedPwHash)
    if err == sql.ErrNoRows {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    } else if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if !utils.CheckPasswordHash(req.Password, storedPwHash) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Login succeed",
        "token":   utils.GenerateToken(req.Username),
    })
}
