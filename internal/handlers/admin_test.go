package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"manticore-imports/internal/domain"
)

type fakeAuthService struct {
	usesCognito bool
	loginFn     func(ctx context.Context, username, password string) (string, string, error)
	validateFn  func(token string) (string, error)
}

func (f *fakeAuthService) UsesCognito() bool { return f.usesCognito }
func (f *fakeAuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	return f.loginFn(ctx, username, password)
}
func (f *fakeAuthService) Validate(token string) (string, error) {
	return f.validateFn(token)
}

func TestAdminHandlerLogin(t *testing.T) {
	h := &AdminHandler{
		Auth: &fakeAuthService{
			loginFn: func(_ context.Context, username, password string) (string, string, error) {
				if username == "err" {
					return "", "", errors.New("signing failed")
				}
				if username != "admin" || password != "secret" {
					return "", "", errors.New("invalid credentials")
				}
				return "token-1", "legacy", nil
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
		listPageFn: func(_ context.Context, params domain.ListRequestsParams) (*domain.PaginatedRequests, error) {
			return &domain.PaginatedRequests{Items: []domain.Request{{RequestID: "req-1"}}, HasMore: false}, nil
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
		Metrics: &fakeMetricsService{
			snapshotFn: func(_ context.Context) (*domain.MetricsSnapshot, error) {
				return &domain.MetricsSnapshot{TotalRequests: 1, ByStatus: map[string]int{"NEW": 1}}, nil
			},
		},
		Auth: &fakeAuthService{
			validateFn: func(token string) (string, error) {
				if token == "bad" {
					return "", errors.New("invalid token")
				}
				return "admin-user", nil
			},
			loginFn: func(_ context.Context, _, _ string) (string, string, error) {
				return "unused", "legacy", nil
			},
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
		if !strings.Contains(resp.Body, `"hasMore":false`) {
			t.Fatalf("expected paginated response body=%s", resp.Body)
		}
	})

	t.Run("GetMetrics success", func(t *testing.T) {
		resp, err := h.GetMetrics(events.APIGatewayV2HTTPRequest{
			Headers: map[string]string{"Authorization": "Bearer ok"},
		})
		if err != nil {
			t.Fatalf("GetMetrics returned error: %v", err)
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

type fakeMetricsService struct {
	snapshotFn func(ctx context.Context) (*domain.MetricsSnapshot, error)
}

func (f *fakeMetricsService) Snapshot(ctx context.Context) (*domain.MetricsSnapshot, error) {
	return f.snapshotFn(ctx)
}
