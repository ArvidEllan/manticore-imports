package handlers

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"manticore-imports/internal/domain"
)

type fakeDealScanner struct {
	scanFn func(ctx context.Context, input domain.ScanDealsRequest) (*domain.ScanDealsResult, error)
}

func (f *fakeDealScanner) Scan(ctx context.Context, input domain.ScanDealsRequest) (*domain.ScanDealsResult, error) {
	return f.scanFn(ctx, input)
}

func TestPublicHandlerScanDeals(t *testing.T) {
	h := &PublicHandler{
		DealScanner: &fakeDealScanner{
			scanFn: func(_ context.Context, input domain.ScanDealsRequest) (*domain.ScanDealsResult, error) {
				if input.Query != "Widget" {
					t.Fatalf("unexpected query: %s", input.Query)
				}
				return &domain.ScanDealsResult{
					Query:    input.Query,
					Listings: []domain.ProductListing{{SiteID: "alibaba", TotalUSD: 99.5}},
					BestDeal: &domain.ProductListing{SiteID: "alibaba", TotalUSD: 99.5},
				}, nil
			},
		},
	}

	resp, err := h.ScanDeals(events.APIGatewayV2HTTPRequest{
		Body: `{"query":"Widget"}`,
	})
	if err != nil {
		t.Fatalf("ScanDeals returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
}

func TestPublicHandlerScanDealsValidation(t *testing.T) {
	h := &PublicHandler{
		DealScanner: &fakeDealScanner{
			scanFn: func(_ context.Context, _ domain.ScanDealsRequest) (*domain.ScanDealsResult, error) {
				return nil, nil
			},
		},
	}

	resp, err := h.ScanDeals(events.APIGatewayV2HTTPRequest{Body: `{}`})
	if err != nil {
		t.Fatalf("ScanDeals returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestPublicHandlerScanDealsNotConfigured(t *testing.T) {
	h := &PublicHandler{}
	resp, err := h.ScanDeals(events.APIGatewayV2HTTPRequest{Body: `{"query":"Widget"}`})
	if err != nil {
		t.Fatalf("ScanDeals returned error: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
}
