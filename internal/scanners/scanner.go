package scanners

import (
	"context"

	"manticore-imports/internal/domain"
)

type Scanner interface {
	SiteID() string
	SiteName() string
	Country() string
	Currency() string
	Search(ctx context.Context, query string) ([]domain.ProductListing, error)
}
