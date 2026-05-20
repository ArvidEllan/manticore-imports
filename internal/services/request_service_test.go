package services

import (
	"context"
	"errors"
	"testing"

	"manticore-imports/internal/domain"
)

type fakeRequestRepo struct {
	createFn       func(ctx context.Context, item domain.Request) error
	getByRefFn     func(ctx context.Context, reference string) (*domain.Request, error)
	listFn         func(ctx context.Context) ([]domain.Request, error)
	listPageFn     func(ctx context.Context, params domain.ListRequestsParams) (*domain.PaginatedRequests, error)
	countByStatusFn func(ctx context.Context) (map[string]int, error)
	getByIDFn      func(ctx context.Context, requestID string) (*domain.Request, error)
	updateStatusFn func(ctx context.Context, requestID, status string, updatedAt string) error
}

func (f *fakeRequestRepo) Create(ctx context.Context, item domain.Request) error {
	return f.createFn(ctx, item)
}
func (f *fakeRequestRepo) GetByReference(ctx context.Context, reference string) (*domain.Request, error) {
	return f.getByRefFn(ctx, reference)
}
func (f *fakeRequestRepo) List(ctx context.Context) ([]domain.Request, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}
func (f *fakeRequestRepo) ListPage(ctx context.Context, params domain.ListRequestsParams) (*domain.PaginatedRequests, error) {
	if f.listPageFn == nil {
		return &domain.PaginatedRequests{}, nil
	}
	return f.listPageFn(ctx, params)
}
func (f *fakeRequestRepo) CountByStatus(ctx context.Context) (map[string]int, error) {
	if f.countByStatusFn == nil {
		return map[string]int{}, nil
	}
	return f.countByStatusFn(ctx)
}
func (f *fakeRequestRepo) GetByID(ctx context.Context, requestID string) (*domain.Request, error) {
	if f.getByIDFn == nil {
		return nil, nil
	}
	return f.getByIDFn(ctx, requestID)
}
func (f *fakeRequestRepo) UpdateStatus(ctx context.Context, requestID, status string, updatedAt string) error {
	return f.updateStatusFn(ctx, requestID, status, updatedAt)
}

type fakeAuditRepo struct {
	createFn func(ctx context.Context, event domain.AuditEvent) error
}

func (f *fakeAuditRepo) Create(ctx context.Context, event domain.AuditEvent) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, event)
}

type fakeEmailSender struct {
	sendFn     func(ctx context.Context, to, subject, body string) error
	sendHTMLFn func(ctx context.Context, to, subject, htmlBody, textBody string) error
}

func (f *fakeEmailSender) Send(ctx context.Context, to, subject, body string) error {
	if f.sendFn == nil {
		return nil
	}
	return f.sendFn(ctx, to, subject, body)
}
func (f *fakeEmailSender) SendHTML(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if f.sendHTMLFn != nil {
		return f.sendHTMLFn(ctx, to, subject, htmlBody, textBody)
	}
	if f.sendFn != nil {
		return f.sendFn(ctx, to, subject, textBody)
	}
	return nil
}

func TestRequestServiceCreateQuote(t *testing.T) {
	var created domain.Request
	requests := &fakeRequestRepo{
		createFn: func(_ context.Context, item domain.Request) error {
			created = item
			return nil
		},
	}
	var auditCreated bool
	audit := &fakeAuditRepo{
		createFn: func(_ context.Context, event domain.AuditEvent) error {
			if event.EventType != "REQUEST_CREATED" {
				t.Fatalf("unexpected audit event type: %s", event.EventType)
			}
			auditCreated = true
			return nil
		},
	}
	var emailSent bool
	email := &fakeEmailSender{
		sendHTMLFn: func(_ context.Context, to, subject, htmlBody, textBody string) error {
			if to == "" || subject == "" || htmlBody == "" || textBody == "" {
				t.Fatalf("expected non-empty html email payload")
			}
			emailSent = true
			return nil
		},
	}
	svc := NewRequestService(requests, audit, email)

	item, err := svc.CreateQuote(context.Background(), domain.CreateQuoteRequest{
		CustomerName:    "Jane Doe",
		Email:           "jane@example.com",
		Phone:           "123",
		ProductName:     "Widget",
		ProductCategory: "Tools",
		Quantity:        2,
		SourceCountry:   "KE",
	})
	if err != nil {
		t.Fatalf("CreateQuote returned error: %v", err)
	}
	if item == nil {
		t.Fatalf("expected non-nil request")
	}
	if item.RequestID == "" || item.Reference == "" {
		t.Fatalf("expected generated ids: %+v", item)
	}
	if created.RequestID == "" || created.Status != domain.StatusNew {
		t.Fatalf("request not persisted as expected: %+v", created)
	}
	if !auditCreated {
		t.Fatalf("expected audit event to be created")
	}
	if !emailSent {
		t.Fatalf("expected email to be sent")
	}
}

func TestRequestServiceLookupByReferenceAndEmail(t *testing.T) {
	svc := NewRequestService(
		&fakeRequestRepo{
			getByRefFn: func(_ context.Context, reference string) (*domain.Request, error) {
				if reference == "err" {
					return nil, errors.New("query error")
				}
				return &domain.Request{Reference: reference, Email: "a@example.com"}, nil
			},
		},
		&fakeAuditRepo{},
		&fakeEmailSender{},
	)

	item, err := svc.LookupByReferenceAndEmail(context.Background(), "ref-1", "a@example.com")
	if err != nil || item == nil {
		t.Fatalf("expected successful lookup, got item=%v err=%v", item, err)
	}

	item, err = svc.LookupByReferenceAndEmail(context.Background(), "ref-1", "b@example.com")
	if err != nil {
		t.Fatalf("expected nil,nil on email mismatch, got err=%v", err)
	}
	if item != nil {
		t.Fatalf("expected nil item on email mismatch")
	}

	_, err = svc.LookupByReferenceAndEmail(context.Background(), "err", "a@example.com")
	if err == nil {
		t.Fatalf("expected error from repository")
	}
}

func TestRequestServiceUpdateStatus(t *testing.T) {
	var updated bool
	var audited bool
	svc := NewRequestService(
		&fakeRequestRepo{
			updateStatusFn: func(_ context.Context, requestID, status string, updatedAt string) error {
				updated = requestID == "req-1" && status == domain.StatusUnderReview && updatedAt != ""
				return nil
			},
		},
		&fakeAuditRepo{
			createFn: func(_ context.Context, event domain.AuditEvent) error {
				audited = event.EventType == "STATUS_UPDATED" && event.Actor == "admin"
				return nil
			},
		},
		&fakeEmailSender{},
	)

	if err := svc.UpdateStatus(context.Background(), "req-1", domain.StatusUnderReview, "admin"); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	if !updated || !audited {
		t.Fatalf("expected update and audit to execute (updated=%v audited=%v)", updated, audited)
	}
}
