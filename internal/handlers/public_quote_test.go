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

type fakeRequestService struct {
	createQuoteFn               func(ctx context.Context, input domain.CreateQuoteRequest) (*domain.Request, error)
	lookupByReferenceAndEmailFn func(ctx context.Context, reference, email string) (*domain.Request, error)
	listFn                      func(ctx context.Context) ([]domain.Request, error)
	listPageFn                  func(ctx context.Context, params domain.ListRequestsParams) (*domain.PaginatedRequests, error)
	getByIDFn                   func(ctx context.Context, id string) (*domain.Request, error)
	updateStatusFn              func(ctx context.Context, requestID, status, actor string) error
}

func (f *fakeRequestService) CreateQuote(ctx context.Context, input domain.CreateQuoteRequest) (*domain.Request, error) {
	return f.createQuoteFn(ctx, input)
}
func (f *fakeRequestService) LookupByReferenceAndEmail(ctx context.Context, reference, email string) (*domain.Request, error) {
	return f.lookupByReferenceAndEmailFn(ctx, reference, email)
}
func (f *fakeRequestService) List(ctx context.Context) ([]domain.Request, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}
func (f *fakeRequestService) ListPage(ctx context.Context, params domain.ListRequestsParams) (*domain.PaginatedRequests, error) {
	if f.listPageFn == nil {
		return &domain.PaginatedRequests{}, nil
	}
	return f.listPageFn(ctx, params)
}
func (f *fakeRequestService) GetByID(ctx context.Context, id string) (*domain.Request, error) {
	if f.getByIDFn == nil {
		return nil, nil
	}
	return f.getByIDFn(ctx, id)
}
func (f *fakeRequestService) UpdateStatus(ctx context.Context, requestID, status, actor string) error {
	if f.updateStatusFn == nil {
		return nil
	}
	return f.updateStatusFn(ctx, requestID, status, actor)
}

type fakeUploadService struct {
	createPresignedUploadFn func(ctx context.Context, requestID, fileName, contentType string) (string, string, error)
}

func (f *fakeUploadService) CreatePresignedUpload(ctx context.Context, requestID, fileName, contentType string) (string, string, error) {
	return f.createPresignedUploadFn(ctx, requestID, fileName, contentType)
}

func TestPublicHandlerCreateQuote(t *testing.T) {
	h := &PublicHandler{
		Requests: &fakeRequestService{
			createQuoteFn: func(_ context.Context, _ domain.CreateQuoteRequest) (*domain.Request, error) {
				return &domain.Request{RequestID: "req-1", Reference: "MANT-20260508-ABCDE", Status: domain.StatusNew}, nil
			},
		},
	}

	resp, err := h.CreateQuote(events.APIGatewayV2HTTPRequest{
		Body: `{"customerName":"A","email":"a@example.com","phone":"123","productName":"P","productCategory":"Cat","quantity":1,"sourceCountry":"KE"}`,
	})
	if err != nil {
		t.Fatalf("CreateQuote returned error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, resp.StatusCode, resp.Body)
	}
	if !strings.Contains(resp.Body, `"requestId":"req-1"`) {
		t.Fatalf("expected response to include requestId, got body=%s", resp.Body)
	}
}

func TestPublicHandlerCreateQuoteValidationAndErrors(t *testing.T) {
	h := &PublicHandler{
		Requests: &fakeRequestService{
			createQuoteFn: func(_ context.Context, _ domain.CreateQuoteRequest) (*domain.Request, error) {
				return nil, errors.New("boom")
			},
		},
	}

	cases := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "invalid json", body: `{"bad"`, wantStatus: http.StatusBadRequest},
		{name: "validation error", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"customerName":"A","email":"a@example.com","phone":"123","productName":"P","productCategory":"Cat","quantity":1,"sourceCountry":"KE"}`, wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := h.CreateQuote(events.APIGatewayV2HTTPRequest{Body: tc.body})
			if err != nil {
				t.Fatalf("CreateQuote returned error: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d body=%s", tc.wantStatus, resp.StatusCode, resp.Body)
			}
		})
	}
}

func TestPublicHandlerGetStatus(t *testing.T) {
	h := &PublicHandler{
		Requests: &fakeRequestService{
			lookupByReferenceAndEmailFn: func(_ context.Context, reference, email string) (*domain.Request, error) {
				if reference == "missing" {
					return nil, nil
				}
				if reference == "err" {
					return nil, errors.New("db down")
				}
				return &domain.Request{RequestID: "req-1", Reference: reference, Email: email}, nil
			},
		},
	}

	cases := []struct {
		name       string
		req        events.APIGatewayV2HTTPRequest
		wantStatus int
	}{
		{
			name: "missing email",
			req: events.APIGatewayV2HTTPRequest{
				PathParameters: map[string]string{"reference": "ref"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			req: events.APIGatewayV2HTTPRequest{
				PathParameters:        map[string]string{"reference": "missing"},
				QueryStringParameters: map[string]string{"email": "a@example.com"},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "service error",
			req: events.APIGatewayV2HTTPRequest{
				PathParameters:        map[string]string{"reference": "err"},
				QueryStringParameters: map[string]string{"email": "a@example.com"},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			req: events.APIGatewayV2HTTPRequest{
				PathParameters:        map[string]string{"reference": "ref"},
				QueryStringParameters: map[string]string{"email": "a@example.com"},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := h.GetStatus(tc.req)
			if err != nil {
				t.Fatalf("GetStatus returned error: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d body=%s", tc.wantStatus, resp.StatusCode, resp.Body)
			}
		})
	}
}

func TestPublicHandlerCreatePresignedUpload(t *testing.T) {
	h := &PublicHandler{
		Uploads: &fakeUploadService{
			createPresignedUploadFn: func(_ context.Context, requestID, fileName, contentType string) (string, string, error) {
				if requestID == "err" {
					return "", "", errors.New("s3 down")
				}
				return "doc-1", "https://example.com/upload", nil
			},
		},
	}

	cases := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "invalid json", body: `{"x"`, wantStatus: http.StatusBadRequest},
		{name: "missing fields", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"requestId":"err","fileName":"x.pdf","contentType":"application/pdf"}`, wantStatus: http.StatusInternalServerError},
		{name: "success", body: `{"requestId":"req-1","fileName":"x.pdf","contentType":"application/pdf"}`, wantStatus: http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := h.CreatePresignedUpload(events.APIGatewayV2HTTPRequest{Body: tc.body})
			if err != nil {
				t.Fatalf("CreatePresignedUpload returned error: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d body=%s", tc.wantStatus, resp.StatusCode, resp.Body)
			}
		})
	}
}
