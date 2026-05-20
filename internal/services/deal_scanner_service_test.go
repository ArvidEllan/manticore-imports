package services

import (
	"context"
	"testing"

	"manticore-imports/internal/domain"
	"manticore-imports/internal/scanners"
	"manticore-imports/internal/scanners/sites"
)

type stubScanner struct {
	id      string
	country string
	price   float64
	err     error
}

func (s *stubScanner) SiteID() string   { return s.id }
func (s *stubScanner) SiteName() string { return s.id }
func (s *stubScanner) Country() string  { return s.country }
func (s *stubScanner) Currency() string { return "USD" }
func (s *stubScanner) Search(_ context.Context, query string) ([]domain.ProductListing, error) {
	if s.err != nil {
		return nil, s.err
	}
	return []domain.ProductListing{{
		SiteID:      s.id,
		SiteName:    s.id,
		Country:     s.country,
		Currency:    "USD",
		ProductName: query,
		Price:       s.price,
		PriceUSD:    s.price,
		InStock:     true,
	}}, nil
}

func TestDealScannerServiceScanRanksBestDeal(t *testing.T) {
	svc := NewDealScannerService([]scanners.Scanner{
		&stubScanner{id: "expensive", country: "US", price: 500},
		&stubScanner{id: "cheap", country: "CN", price: 100},
	})

	result, err := svc.Scan(context.Background(), domain.ScanDealsRequest{
		Query: "Industrial pump",
	})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if result.BestDeal == nil {
		t.Fatalf("expected best deal")
	}
	if result.BestDeal.SiteID != "cheap" {
		t.Fatalf("expected cheapest listing, got %s", result.BestDeal.SiteID)
	}
	if len(result.Listings) != 2 {
		t.Fatalf("expected 2 listings, got %d", len(result.Listings))
	}
	if result.Listings[0].SiteID != "cheap" {
		t.Fatalf("expected listings sorted by total price")
	}
}

func TestDealScannerServiceScanFiltersByCountry(t *testing.T) {
	svc := NewDealScannerService([]scanners.Scanner{
		&stubScanner{id: "us-site", country: "US", price: 200},
		&stubScanner{id: "cn-site", country: "CN", price: 150},
	})

	result, err := svc.Scan(context.Background(), domain.ScanDealsRequest{
		Query:     "Laptop stand",
		Countries: []string{"CN"},
	})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(result.Listings) != 1 || result.Listings[0].SiteID != "cn-site" {
		t.Fatalf("expected only CN listing, got %+v", result.Listings)
	}
}

func TestDealScannerServiceScanFromProductURL(t *testing.T) {
	svc := NewDealScannerService([]scanners.Scanner{
		sites.NewAlibaba(),
	})

	result, err := svc.Scan(context.Background(), domain.ScanDealsRequest{
		ProductURL: "https://www.alibaba.com/product-detail/industrial-valve-123.html",
	})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if result.Query == "" {
		t.Fatalf("expected query derived from product URL")
	}
	if len(result.Listings) == 0 {
		t.Fatalf("expected at least one listing")
	}
}

func TestDealScannerServiceDefaultRegistry(t *testing.T) {
	svc := NewDealScannerService(scanners.DefaultRegistry())
	result, err := svc.Scan(context.Background(), domain.ScanDealsRequest{
		Query: "Bluetooth speaker",
	})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(result.ScannedSites) < 5 {
		t.Fatalf("expected multiple sites scanned, got %d", len(result.ScannedSites))
	}
	if result.BestDeal == nil {
		t.Fatalf("expected best deal from registry scan")
	}
}

func TestDealScannerServiceRequiresQueryOrURL(t *testing.T) {
	svc := NewDealScannerService(scanners.DefaultRegistry())
	_, err := svc.Scan(context.Background(), domain.ScanDealsRequest{})
	if err == nil {
		t.Fatalf("expected error for empty input")
	}
}
