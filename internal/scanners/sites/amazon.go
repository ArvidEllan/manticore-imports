package sites

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"manticore-imports/internal/domain"
)

type AmazonScanner struct {
	id       string
	name     string
	country  string
	currency string
	domain   string
}

func NewAmazon(id, name, country, currency, domain string) *AmazonScanner {
	return &AmazonScanner{id: id, name: name, country: country, currency: currency, domain: domain}
}

func (s *AmazonScanner) SiteID() string   { return s.id }
func (s *AmazonScanner) SiteName() string { return s.name }
func (s *AmazonScanner) Country() string  { return s.country }
func (s *AmazonScanner) Currency() string { return s.currency }

func (s *AmazonScanner) Search(ctx context.Context, query string) ([]domain.ProductListing, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	baseUSD := basePriceUSD(q)
	multiplier := 1.0
	switch s.country {
	case "GB":
		multiplier = 0.92
	case "DE":
		multiplier = 0.88
	}
	priceUSD := applySiteMultiplier(baseUSD, multiplier)
	encoded := url.QueryEscape(q)
	productURL := fmt.Sprintf("https://www.%s/s?k=%s", s.domain, encoded)
	return []domain.ProductListing{{
		SiteID:       s.id,
		SiteName:     s.name,
		Country:      s.country,
		Currency:     s.currency,
		ProductName:  q,
		ProductURL:   productURL,
		InStock:      true,
		SellerRating: sellerRating(s.id, q),
		Price:        priceUSD,
		PriceUSD:     priceUSD,
	}}, nil
}
