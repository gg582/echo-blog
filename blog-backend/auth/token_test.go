package auth

import (
	"strings"
	"testing"
	"time"
)

func mustSigner(t *testing.T, secret string) *Signer {
	t.Helper()
	s, err := NewSigner(secret)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTokenRoundTrip(t *testing.T) {
	s := mustSigner(t, "test-secret")
	token := s.Issue("admin")
	user, ok := s.Verify(token)
	if !ok {
		t.Fatalf("expected issued token to be valid, got %q", token)
	}
	if user != "admin" {
		t.Fatalf("expected username admin, got %q", user)
	}
	if !strings.HasPrefix(token, "admin.") {
		t.Fatalf("expected token to start with username, got %q", token)
	}
}

func TestTokenWrongSecretRejected(t *testing.T) {
	token := mustSigner(t, "secret-a").Issue("admin")
	if _, ok := mustSigner(t, "secret-b").Verify(token); ok {
		t.Fatal("expected token signed with a different secret to be rejected")
	}
}

func TestTokenTamperedRejected(t *testing.T) {
	s := mustSigner(t, "test-secret")
	parts := strings.Split(s.Issue("admin"), ".")
	if _, ok := s.Verify(parts[0] + "." + parts[1] + ".deadbeef"); ok {
		t.Fatal("expected tampered token to be rejected")
	}
	if _, ok := s.Verify("other." + parts[1] + "." + parts[2]); ok {
		t.Fatal("expected token with swapped username to be rejected")
	}
	if _, ok := s.Verify("garbage"); ok {
		t.Fatal("expected malformed token to be rejected")
	}
}

func TestTokenExpiry(t *testing.T) {
	s := mustSigner(t, "test-secret")
	token := s.Issue("admin")

	s.now = func() time.Time { return time.Now().Add(tokenTTL + time.Minute) }
	if _, ok := s.Verify(token); ok {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestRandomSecretWhenEmpty(t *testing.T) {
	a, b := mustSigner(t, ""), mustSigner(t, "")
	if _, ok := b.Verify(a.Issue("admin")); ok {
		t.Fatal("expected two random secrets to differ")
	}
}
