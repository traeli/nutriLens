package auth

import (
	"testing"
	"time"
)

func TestTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("01234567890123456789012345678901", time.Hour)
	token, err := manager.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	userID, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if userID != 42 {
		t.Fatalf("Parse() userID = %d, want 42", userID)
	}
}

func TestTokenRejectsDifferentSecret(t *testing.T) {
	issuer := NewTokenManager("01234567890123456789012345678901", time.Hour)
	parser := NewTokenManager("abcdefghijklmnopqrstuvwxyz123456", time.Hour)
	token, err := issuer.Issue(7)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if _, err := parser.Parse(token); err == nil {
		t.Fatal("Parse() unexpectedly accepted a token signed with another secret")
	}
}
