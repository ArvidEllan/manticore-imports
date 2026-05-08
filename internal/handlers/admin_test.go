package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"manticore-imports/internal/domain"
)

type fakeTokenService struct {
	generateFn func(subject string, ttl time.Duration) (string, error)
	validateFn func(token string) (string, error)
}

func (f *fakeTokenService) Generate(subject string, ttl time.Duration) (string, error) {
	return f.generateFn(subject, ttl)
}
func (f *fakeTokenService) Validate(token string) (string, error) {
	return f.validateFn(token)
}

func TestAdminHandlerLogin(t *testing.T) {
	h := &AdminHandler{
		AdminUsername: "admin",
		AdminPassword: "secret",
		TokenService: &fakeTokenService{
			generateFn: func(subject string, ttl time.Duration) (string, error) {
				if subject == "err" {
					return "", errors.New("signing failed")
				}
				if ttl != 12*time.Hour {
					t.Fatalf("unexpected ttl: %s", ttl)
				}
				return "token-1", nil
			},
		},
	}

	cases := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "invalid json", body: `{"x"`, wantStatus: http.StatusBadRequest},
		{name: "invalid credentials", body: `{"username":"admin","password":"bad"}`, wantStatus: http.StatusUnauthorized},
		{name: "success", body: `{"username":"admin","password":"secret"}`, wantStatus: http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := h.Login(events.APIGatewayV2HTTPRequest{Body: tc.body})
			if err != nil {
				t.Fatalf("Login returned error: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d body=%s", tc.wantStatus, resp.StatusCode, resp.Body)
			}
		})
	}
}

func TestAdminHandlerAuthorizedEndpoints(t *testing.T) {
	requests := &fakeRequestService{
		listFn: func(_ context.Context) ([]domain.Request, error) {
			return []domain.Request{{RequestID: "req-1"}}, nil
		},
		getByIDFn: func(_ context.Context, id string) (*domain.Request, error) {
			if id == "missing" {
				return nil, nil
			}
			if id == "err" {
				return nil, errors.New("db error")
			}
			return &domain.Request{RequestID: id}, nil
		},
		updateStatusFn: func(_ context.Context, requestID, status, actor string) error {
			if requestID == "err" {
				return errors.New("update failed")
			}
			if actor == "" {
				t.Fatalf("expected actor subject from token")
			}
			return nil
		},
	}
	h := &AdminHandler{
		Requests: requests,
		TokenService: &fakeTokenService{
			validateFn: func(token string) (string, error) {
				if token == "bad" {
					return "", errors.New("invalid token")
				}
				return "admin-user", nil
			},
			generateFn: func(subject string, ttl time.Duration) (string, error) { return "unused", nil },
		},
	}

	t.Run("ListRequests unauthorized", func(t *testing.T) {
		resp, err := h.ListRequests(events.APIGatewayV2HTTPRequest{})
		if err != nil {
			t.Fatalf("ListRequests returned error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("ListRequests success", func(t *testing.T) {
		resp, err := h.ListRequests(events.APIGatewayV2HTTPRequest{
			Headers: map[string]string{"Authorization": "Bearer ok"},
		})
		if err != nil {
			t.Fatalf("ListRequests returned error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resp.StatusCode, resp.Body)
		}
	})

	t.Run("GetRequest not found", func(t *testing.T) {
		resp, err := h.GetRequest(events.APIGatewayV2HTTPRequest{
			Headers:        map[string]string{"Authorization": "Bearer ok"},
			PathParameters: map[string]string{"id": "missing"},
		})
		if err != nil {
			t.Fatalf("GetRequest returned error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, resp.StatusCode, resp.Body)
		}
	})

	t.Run("GetRequest service error", func(t *testing.T) {
		resp, err := h.GetRequest(events.APIGatewayV2HTTPRequest{
			Headers:        map[string]string{"Authorization": "Bearer ok"},
			PathParameters: map[string]string{"id": "err"},
		})
		if err != nil {
			t.Fatalf("GetRequest returned error: %v", err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, resp.StatusCode, resp.Body)
		}
	})

	t.Run("UpdateStatus invalid status", func(t *testing.T) {
		resp, err := h.UpdateStatus(events.APIGatewayV2HTTPRequest{
			Headers:        map[string]string{"Authorization": "Bearer ok"},
			PathParameters: map[string]string{"id": "req-1"},
			Body:           `{"status":"NOT_A_STATUS"}`,
		})
		if err != nil {
			t.Fatalf("UpdateStatus returned error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, resp.StatusCode, resp.Body)
		}
	})

	t.Run("UpdateStatus success", func(t *testing.T) {
		resp, err := h.UpdateStatus(events.APIGatewayV2HTTPRequest{
			Headers:        map[string]string{"authorization": "Bearer ok"},
			PathParameters: map[string]string{"id": "req-1"},
			Body:           `{"status":"UNDER_REVIEW"}`,
		})
		if err != nil {
			t.Fatalf("UpdateStatus returned error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resp.StatusCode, resp.Body)
		}
		if !strings.Contains(resp.Body, "status updated") {
			t.Fatalf("unexpected response body: %s", resp.Body)
		}
	})
}
