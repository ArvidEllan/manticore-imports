package sites

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"manticore-imports/internal/domain"
)

type AliExpressScanner struct{}

func NewAliExpress() *AliExpressScanner { return &AliExpressScanner{} }

func (s *AliExpressScanner) SiteID() string   { return "aliexpress" }
func (s *AliExpressScanner) SiteName() string { return "AliExpress" }
func (s *AliExpressScanner) Country() string  { return "CN" }
func (s *AliExpressScanner) Currency() string { return "USD" }

func (s *AliExpressScanner) Search(ctx context.Context, query string) ([]domain.ProductListing, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	priceUSD := applySiteMultiplier(basePriceUSD(q), 0.72)
	encoded := url.QueryEscape(q)
	return []domain.ProductListing{{
		SiteID:       s.SiteID(),
		SiteName:     s.SiteName(),
		Country:      s.Country(),
		Currency:     s.Currency(),
		ProductName:  q,
		ProductURL:   fmt.Sprintf("https://www.aliexpress.com/wholesale?SearchText=%s", encoded),
		InStock:      true,
		SellerRating: sellerRating(s.SiteID(), q),
		Price:        priceUSD,
		PriceUSD:     priceUSD,
	}}, nil
}
