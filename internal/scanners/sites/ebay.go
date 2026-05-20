package sites

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"manticore-imports/internal/domain"
)

type EbayScanner struct {
	id       string
	name     string
	country  string
	currency string
	domain   string
}

func NewEbay(id, name, country, currency, domain string) *EbayScanner {
	return &EbayScanner{id: id, name: name, country: country, currency: currency, domain: domain}
}

func (s *EbayScanner) SiteID() string   { return s.id }
func (s *EbayScanner) SiteName() string { return s.name }
func (s *EbayScanner) Country() string  { return s.country }
func (s *EbayScanner) Currency() string { return s.currency }

func (s *EbayScanner) Search(ctx context.Context, query string) ([]domain.ProductListing, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	multiplier := 0.85
	if s.country == "GB" {
		multiplier = 0.87
	}
	priceUSD := applySiteMultiplier(basePriceUSD(q), multiplier)
	encoded := url.QueryEscape(q)
	return []domain.ProductListing{{
		SiteID:       s.id,
		SiteName:     s.name,
		Country:      s.country,
		Currency:     s.currency,
		ProductName:  q + " (used/new listings)",
		ProductURL:   fmt.Sprintf("https://www.%s/sch/i.html?_nkw=%s", s.domain, encoded),
		InStock:      true,
		SellerRating: sellerRating(s.id, q),
		Price:        priceUSD,
		PriceUSD:     priceUSD,
	}}, nil
}
