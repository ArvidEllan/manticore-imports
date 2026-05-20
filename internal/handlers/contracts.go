package handlers

import (
	"context"

	"manticore-imports/internal/domain"
)

type requestService interface {
	CreateQuote(ctx context.Context, input domain.CreateQuoteRequest) (*domain.Request, error)
	LookupByReferenceAndEmail(ctx context.Context, reference, email string) (*domain.Request, error)
	List(ctx context.Context) ([]domain.Request, error)
	ListPage(ctx context.Context, params domain.ListRequestsParams) (*domain.PaginatedRequests, error)
	GetByID(ctx context.Context, id string) (*domain.Request, error)
	UpdateStatus(ctx context.Context, requestID, status, actor string) error
}

type uploadService interface {
	CreatePresignedUpload(ctx context.Context, requestID, fileName, contentType string) (string, string, error)
}

type authService interface {
	UsesCognito() bool
	Login(ctx context.Context, username, password string) (token string, authType string, err error)
	Validate(token string) (string, error)
}

type metricsService interface {
	Snapshot(ctx context.Context) (*domain.MetricsSnapshot, error)
}

type dealScannerService interface {
	Scan(ctx context.Context, input domain.ScanDealsRequest) (*domain.ScanDealsResult, error)
}
