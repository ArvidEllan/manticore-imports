package services

import (
	"context"
	"testing"
	"time"

	"manticore-imports/internal/domain"
)

func TestAuthServiceLegacyLogin(t *testing.T) {
	svc := NewAuthService(nil, NewTokenService("secret"), "admin", "pass")

	token, authType, err := svc.Login(context.Background(), "admin", "pass")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if token == "" || authType != "legacy" {
		t.Fatalf("unexpected login response token=%q authType=%q", token, authType)
	}

	subject, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if subject != "admin" {
		t.Fatalf("expected subject admin, got %s", subject)
	}
}

func TestAuthServiceLegacyInvalidCredentials(t *testing.T) {
	svc := NewAuthService(nil, NewTokenService("secret"), "admin", "pass")
	_, _, err := svc.Login(context.Background(), "admin", "wrong")
	if err == nil {
		t.Fatalf("expected login error")
	}
}

func TestAuthServiceUsesCognitoWhenEnabled(t *testing.T) {
	cognito := &CognitoAuthService{userPoolID: "pool", clientID: "client", region: "eu-west-1"}
	svc := NewAuthService(cognito, NewTokenService("secret"), "admin", "pass")
	if !svc.UsesCognito() {
		t.Fatalf("expected cognito to be enabled")
	}
}

func TestAuthServiceLegacyTokenExpiry(t *testing.T) {
	tokenSvc := NewTokenService("secret")
	token, err := tokenSvc.Generate("admin", -1*time.Hour)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if _, err := tokenSvc.Validate(token); err == nil {
		t.Fatalf("expected expired token error")
	}
}

func TestMetricsServiceSnapshot(t *testing.T) {
	svc := NewMetricsService(&fakeRequestRepo{
		countByStatusFn: func(_ context.Context) (map[string]int, error) {
			return map[string]int{domain.StatusNew: 3, domain.StatusCompleted: 1}, nil
		},
	})
	snapshot, err := svc.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if snapshot.TotalRequests != 4 {
		t.Fatalf("expected total 4, got %d", snapshot.TotalRequests)
	}
}
