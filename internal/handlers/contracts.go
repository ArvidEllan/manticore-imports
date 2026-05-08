package handlers

import (
	"context"
	"time"

	"manticore-imports/internal/domain"
)

type requestService interface {
	CreateQuote(ctx context.Context, input domain.CreateQuoteRequest) (*domain.Request, error)
	LookupByReferenceAndEmail(ctx context.Context, reference, email string) (*domain.Request, error)
	List(ctx context.Context) ([]domain.Request, error)
	GetByID(ctx context.Context, id string) (*domain.Request, error)
	UpdateStatus(ctx context.Context, requestID, status, actor string) error
}

type uploadService interface {
	CreatePresignedUpload(ctx context.Context, requestID, fileName, contentType string) (string, string, error)
}

type tokenService interface {
	Generate(subject string, ttl time.Duration) (string, error)
	Validate(token string) (string, error)
}
