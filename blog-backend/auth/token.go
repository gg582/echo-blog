// Package auth handles the admin account: password hashing, the users table
// and the signed bearer tokens handed out on login.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// tokenTTL is how long an issued token stays valid.
const tokenTTL = 72 * time.Hour

// Signer issues and verifies tokens of the form "<username>.<expiryUnix>.<sig>",
// where sig is the hex HMAC-SHA256 of "<username>.<expiryUnix>".
type Signer struct {
	secret []byte
	now    func() time.Time
}

// NewSigner returns a Signer keyed with secret. If secret is empty, a random
// key is generated and a warning is logged: tokens issued before a restart
// will no longer be valid.
func NewSigner(secret string) (*Signer, error) {
	s := &Signer{secret: []byte(secret), now: time.Now}
	if secret != "" {
		return s, nil
	}
	s.secret = make([]byte, 32)
	if _, err := rand.Read(s.secret); err != nil {
		return nil, fmt.Errorf("generate random auth secret: %w", err)
	}
	log.Print("WARNING: AUTH_SECRET is not set; a random secret was generated. All tokens are invalidated on restart.")
	return s, nil
}

// Issue returns a new token for username.
func (s *Signer) Issue(username string) string {
	expiry := s.now().Add(tokenTTL).Unix()
	payload := fmt.Sprintf("%s.%d", username, expiry)
	return payload + "." + hex.EncodeToString(s.sign(payload))
}

// Verify checks the token signature and expiry and returns the username it
// was issued for.
func (s *Signer) Verify(token string) (username string, ok bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}

	payload := parts[0] + "." + parts[1]
	providedSig, err := hex.DecodeString(parts[2])
	if err != nil || !hmac.Equal(s.sign(payload), providedSig) {
		return "", false
	}

	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || s.now().Unix() >= expiry {
		return "", false
	}
	return parts[0], true
}

func (s *Signer) sign(payload string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return mac.Sum(nil)
}
