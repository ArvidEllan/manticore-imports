package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manticore-imports/internal/domain"
)

type requestRepository interface {
	Create(ctx context.Context, item domain.Request) error
	GetByReference(ctx context.Context, reference string) (*domain.Request, error)
	List(ctx context.Context) ([]domain.Request, error)
	GetByID(ctx context.Context, requestID string) (*domain.Request, error)
	UpdateStatus(ctx context.Context, requestID, status string, updatedAt string) error
}

type auditRepository interface {
	Create(ctx context.Context, event domain.AuditEvent) error
}

type emailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type RequestService struct {
	requests requestRepository
	audit    auditRepository
	email    emailSender
}

func NewRequestService(requests requestRepository, audit auditRepository, email emailSender) *RequestService {
	return &RequestService{requests: requests, audit: audit, email: email}
}

func (s *RequestService) CreateQuote(ctx context.Context, input domain.CreateQuoteRequest) (*domain.Request, error) {
	now := time.Now().UTC()
	requestID := uuid.NewString()
	ref := GenerateReference(now)
	item := domain.Request{
		PK:                "REQUEST#" + requestID,
		RequestID:         requestID,
		Reference:         ref,
		CustomerName:      input.CustomerName,
		Email:             input.Email,
		Phone:             input.Phone,
		CompanyName:       input.CompanyName,
		ProductName:       input.ProductName,
		ProductCategory:   input.ProductCategory,
		Quantity:          input.Quantity,
		SourceCountry:     input.SourceCountry,
		PreferredTimeline: input.PreferredTimeline,
		ProductURL:        input.ProductURL,
		Notes:             input.Notes,
		Status:            domain.StatusNew,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.requests.Create(ctx, item); err != nil { return nil, err }
	_ = s.audit.Create(ctx, domain.AuditEvent{
		PK:        "REQUEST#" + requestID,
		SK:        "EVENT#" + now.Format(time.RFC3339Nano),
		EventType: "REQUEST_CREATED",
		Actor:     "public",
		Details:   fmt.Sprintf("request %s created", ref),
		CreatedAt: now,
	})
	_ = s.email.Send(ctx, item.Email, "Manticore quote request received", fmt.Sprintf("Your request %s has been received.", item.Reference))
	return &item, nil
}

func (s *RequestService) LookupByReferenceAndEmail(ctx context.Context, reference, email string) (*domain.Request, error) {
	item, err := s.requests.GetByReference(ctx, reference)
	if err != nil || item == nil { return item, err }
	if item.Email != email {
		return nil, nil
	}
	return item, nil
}

func (s *RequestService) List(ctx context.Context) ([]domain.Request, error) {
	return s.requests.List(ctx)
}

func (s *RequestService) GetByID(ctx context.Context, id string) (*domain.Request, error) {
	return s.requests.GetByID(ctx, id)
}

func (s *RequestService) UpdateStatus(ctx context.Context, requestID, status, actor string) error {
	now := time.Now().UTC()
	if err := s.requests.UpdateStatus(ctx, requestID, status, now.Format(time.RFC3339)); err != nil {
		return err
	}
	return s.audit.Create(ctx, domain.AuditEvent{
		PK:        "REQUEST#" + requestID,
		SK:        "EVENT#" + now.Format(time.RFC3339Nano),
		EventType: "STATUS_UPDATED",
		Actor:     actor,
		Details:   status,
		CreatedAt: now,
	})
}
