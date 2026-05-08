package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"manticore-imports/internal/domain"
	"manticore-imports/internal/handlers"
)

type routeRequestService struct{}

func (r *routeRequestService) CreateQuote(_ context.Context, _ domain.CreateQuoteRequest) (*domain.Request, error) {
	return &domain.Request{RequestID: "req-1"}, nil
}
func (r *routeRequestService) LookupByReferenceAndEmail(_ context.Context, reference, email string) (*domain.Request, error) {
	return &domain.Request{Reference: reference, Email: email}, nil
}
func (r *routeRequestService) List(_ context.Context) ([]domain.Request, error) {
	return []domain.Request{{RequestID: "req-1"}}, nil
}
func (r *routeRequestService) GetByID(_ context.Context, id string) (*domain.Request, error) {
	return &domain.Request{RequestID: id}, nil
}
func (r *routeRequestService) UpdateStatus(_ context.Context, requestID, status, actor string) error {
	return nil
}

type routeUploadService struct{}

func (r *routeUploadService) CreatePresignedUpload(_ context.Context, requestID, fileName, contentType string) (string, string, error) {
	return "doc-1", "https://example.com/upload", nil
}

type routeTokenService struct{}

func (r *routeTokenService) Generate(subject string, ttl time.Duration) (string, error) {
	return "token-1", nil
}
func (r *routeTokenService) Validate(token string) (string, error) {
	return "admin", nil
}

func TestHandlerRouting(t *testing.T) {
	a := &app{
		public: &handlers.PublicHandler{
			Requests: &routeRequestService{},
			Uploads:  &routeUploadService{},
		},
		admin: &handlers.AdminHandler{
			Requests:      &routeRequestService{},
			TokenService:  &routeTokenService{},
			AdminUsername: "admin",
			AdminPassword: "secret",
		},
	}

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		query      map[string]string
		headers    map[string]string
		wantStatus int
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{name: "create quote", method: http.MethodPost, path: "/public/quotes", body: `{"customerName":"A","email":"a@example.com","phone":"123","productName":"P","productCategory":"Cat","quantity":1,"sourceCountry":"KE"}`, wantStatus: http.StatusCreated},
		{name: "public status", method: http.MethodGet, path: "/public/status/REF-1", query: map[string]string{"email": "a@example.com"}, wantStatus: http.StatusOK},
		{name: "presigned upload", method: http.MethodPost, path: "/public/uploads/presign", body: `{"requestId":"req-1","fileName":"x.pdf","contentType":"application/pdf"}`, wantStatus: http.StatusOK},
		{name: "admin login", method: http.MethodPost, path: "/admin/auth/login", body: `{"username":"admin","password":"secret"}`, wantStatus: http.StatusOK},
		{name: "admin list", method: http.MethodGet, path: "/admin/requests", headers: map[string]string{"Authorization": "Bearer token-1"}, wantStatus: http.StatusOK},
		{name: "admin get", method: http.MethodGet, path: "/admin/requests/req-1", headers: map[string]string{"Authorization": "Bearer token-1"}, wantStatus: http.StatusOK},
		{name: "admin update", method: http.MethodPatch, path: "/admin/requests/req-1/status", headers: map[string]string{"Authorization": "Bearer token-1"}, body: `{"status":"UNDER_REVIEW"}`, wantStatus: http.StatusOK},
		{name: "route not found", method: http.MethodGet, path: "/not-found", wantStatus: http.StatusNotFound},
		{name: "invalid status path", method: http.MethodPatch, path: "/admin/requests/x/status/extra", headers: map[string]string{"Authorization": "Bearer token-1"}, body: `{"status":"UNDER_REVIEW"}`, wantStatus: http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := a.handler(context.Background(), events.APIGatewayV2HTTPRequest{
				RawPath: tc.path,
				Body:    tc.body,
				Headers: tc.headers,
				QueryStringParameters: tc.query,
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: tc.method,
					},
				},
			})
			if err != nil {
				t.Fatalf("handler returned error: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d body=%s", tc.wantStatus, resp.StatusCode, resp.Body)
			}
		})
	}
}

func TestExtractRequestIDFromStatusPath(t *testing.T) {
	id, err := extractRequestIDFromStatusPath("/admin/requests/abc/status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "abc" {
		t.Fatalf("expected abc, got %s", id)
	}

	if _, err := extractRequestIDFromStatusPath("/admin/requests/abc/status/extra"); err == nil {
		t.Fatalf("expected error for malformed path")
	}
}
