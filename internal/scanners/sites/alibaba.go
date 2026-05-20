package sites

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"manticore-imports/internal/domain"
)

type AlibabaScanner struct{}

func NewAlibaba() *AlibabaScanner { return &AlibabaScanner{} }

func (s *AlibabaScanner) SiteID() string   { return "alibaba" }
func (s *AlibabaScanner) SiteName() string { return "Alibaba" }
func (s *AlibabaScanner) Country() string  { return "CN" }
func (s *AlibabaScanner) Currency() string { return "USD" }

func (s *AlibabaScanner) Search(ctx context.Context, query string) ([]domain.ProductListing, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	// Wholesale pricing is typically lower than retail marketplaces.
	priceUSD := applySiteMultiplier(basePriceUSD(q), 0.65)
	encoded := url.QueryEscape(q)
	return []domain.ProductListing{{
		SiteID:       s.SiteID(),
		SiteName:     s.SiteName(),
		Country:      s.Country(),
		Currency:     s.Currency(),
		ProductName:  q + " (wholesale MOQ 10)",
		ProductURL:   fmt.Sprintf("https://www.alibaba.com/trade/search?SearchText=%s", encoded),
		InStock:      true,
		SellerRating: sellerRating(s.SiteID(), q),
		Price:        priceUSD,
		PriceUSD:     priceUSD,
	}}, nil
}
