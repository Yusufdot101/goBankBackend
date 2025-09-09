package token

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	id := int64(1)
	ttl := 24 * time.Hour
	scope := "activation"

	token, err := generateToken(id, ttl, scope)
	if err != nil {
		t.Fatalf("expected no errors, got %v", err)
	}

	if token.UserID != id {
		t.Errorf("expected user id %d, got %d", id, token.UserID)
	}

	if token.Plaintext == "" {
		t.Errorf("expected non empty plain text")
	}

	if len(token.hash) != sha256.Size {
		t.Errorf("expected hash size %d, got %d", sha256.Size, len(token.hash))
	}

	// the function will might take some time(<= 1 minute), thats why we don't do
	// time.Until(token.Expiry) !=  ttl
	if time.Until(token.Expiry) < ttl-1*time.Minute {
		t.Error("expected expiry at least ~1 hour from now")
	}

	if time.Until(token.Expiry) > ttl+1*time.Minute {
		t.Error("expiry set too far in the future")
	}

	expectedHash := sha256.Sum256([]byte(token.Plaintext))
	if !bytes.Equal(token.hash, expectedHash[:]) {
		t.Error("hash does not match plaintext")
	}
}
