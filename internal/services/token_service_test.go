package services

import (
	"strings"
	"testing"
	"time"
)

func TestTokenServiceGenerateAndValidate(t *testing.T) {
	svc := NewTokenService("super-secret")

	token, err := svc.Generate("admin", time.Hour)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	sub, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if sub != "admin" {
		t.Fatalf("expected subject admin, got %s", sub)
	}
}

func TestTokenServiceValidateErrors(t *testing.T) {
	svc := NewTokenService("super-secret")
	if _, err := svc.Validate("not-a-jwt"); err == nil {
		t.Fatalf("expected invalid token format error")
	}

	token, err := svc.Generate("admin", time.Hour)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	parts := strings.Split(token, ".")
	parts[1] = "tampered"
	tampered := strings.Join(parts, ".")
	if _, err := svc.Validate(tampered); err == nil {
		t.Fatalf("expected invalid signature error")
	}

	expired, err := svc.Generate("admin", -1*time.Minute)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if _, err := svc.Validate(expired); err == nil {
		t.Fatalf("expected token expired error")
	}
}
