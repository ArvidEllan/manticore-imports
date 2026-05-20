package services

import (
	"context"
	"math"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"manticore-imports/internal/domain"
	"manticore-imports/internal/scanners"
)

var currencyToUSD = map[string]float64{
	"USD": 1.0,
	"GBP": 1.27,
	"EUR": 1.08,
	"CNY": 0.14,
}

type siteScanner interface {
	SiteID() string
	SiteName() string
	Country() string
	Currency() string
	Search(ctx context.Context, query string) ([]domain.ProductListing, error)
}

type DealScannerService struct {
	scanners []siteScanner
	timeout  time.Duration
}

func NewDealScannerService(registry []scanners.Scanner) *DealScannerService {
	adapters := make([]siteScanner, len(registry))
	for i, s := range registry {
		adapters[i] = s
	}
	return &DealScannerService{
		scanners: adapters,
		timeout:  8 * time.Second,
	}
}

func (s *DealScannerService) Scan(ctx context.Context, input domain.ScanDealsRequest) (*domain.ScanDealsResult, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" && strings.TrimSpace(input.ProductURL) != "" {
		query = productNameFromURL(input.ProductURL)
	}
	if query == "" {
		return nil, errEmptyScanQuery
	}

	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	countryFilter := normalizeCountries(input.Countries)
	activeScanners := s.filterScanners(countryFilter, input.ProductURL)

	scanCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var (
		mu       sync.Mutex
		listings []domain.ProductListing
		scanned  []string
		failed   []domain.SiteScanError
		wg       sync.WaitGroup
	)

	for _, scanner := range activeScanners {
		wg.Add(1)
		go func(sc siteScanner) {
			defer wg.Done()
			items, err := sc.Search(scanCtx, query)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failed = append(failed, domain.SiteScanError{SiteID: sc.SiteID(), Error: err.Error()})
				return
			}
			scanned = append(scanned, sc.SiteID())
			for _, item := range items {
				listings = append(listings, normalizeListing(item))
			}
		}(scanner)
	}
	wg.Wait()

	sort.Slice(listings, func(i, j int) bool {
		return listings[i].TotalUSD < listings[j].TotalUSD
	})
	if len(listings) > maxResults {
		listings = listings[:maxResults]
	}

	result := &domain.ScanDealsResult{
		ScanID:       uuid.NewString(),
		Query:        query,
		Listings:     listings,
		ScannedSites: scanned,
		FailedSites:  failed,
		ScannedAt:    time.Now().UTC(),
	}
	if len(listings) > 0 {
		best := listings[0]
		result.BestDeal = &best
	}
	return result, nil
}

func (s *DealScannerService) filterScanners(countryFilter map[string]struct{}, productURL string) []siteScanner {
	if len(countryFilter) == 0 && productURL == "" {
		return s.scanners
	}

	host := ""
	if productURL != "" {
		if u, err := url.Parse(productURL); err == nil {
			host = strings.ToLower(u.Host)
		}
	}

	filtered := make([]siteScanner, 0, len(s.scanners))
	for _, sc := range s.scanners {
		if len(countryFilter) > 0 {
			if _, ok := countryFilter[strings.ToUpper(sc.Country())]; !ok {
				continue
			}
		}
		if host != "" && !siteMatchesHost(sc.SiteID(), host) {
			continue
		}
		filtered = append(filtered, sc)
	}
	if len(filtered) == 0 {
		return s.scanners
	}
	return filtered
}

func siteMatchesHost(siteID, host string) bool {
	host = strings.TrimPrefix(host, "www.")
	switch siteID {
	case "amazon-us":
		return strings.Contains(host, "amazon.com")
	case "amazon-uk":
		return strings.Contains(host, "amazon.co.uk")
	case "amazon-de":
		return strings.Contains(host, "amazon.de")
	case "alibaba":
		return strings.Contains(host, "alibaba.com")
	case "aliexpress":
		return strings.Contains(host, "aliexpress.com")
	case "ebay-us":
		return strings.Contains(host, "ebay.com")
	case "ebay-uk":
		return strings.Contains(host, "ebay.co.uk")
	default:
		return true
	}
}

func normalizeListing(item domain.ProductListing) domain.ProductListing {
	rate, ok := currencyToUSD[strings.ToUpper(item.Currency)]
	if !ok {
		rate = 1.0
	}
	if item.PriceUSD == 0 {
		item.PriceUSD = math.Round(item.Price*rate*100) / 100
	}
	if item.ShippingEstimate == 0 {
		item.ShippingEstimate = shippingEstimateUSD(item.SiteID, item.PriceUSD)
	}
	item.TotalUSD = math.Round((item.PriceUSD+item.ShippingEstimate)*100) / 100
	return item
}

func shippingEstimateUSD(siteID string, priceUSD float64) float64 {
	switch siteID {
	case "alibaba":
		return math.Round((priceUSD*0.08+25)*100) / 100
	case "aliexpress":
		return math.Round((priceUSD*0.05+8)*100) / 100
	case "amazon-us", "amazon-uk", "amazon-de":
		return math.Round((priceUSD*0.03+5)*100) / 100
	default:
		return math.Round((priceUSD*0.04+6)*100) / 100
	}
}

func normalizeCountries(countries []string) map[string]struct{} {
	if len(countries) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(countries))
	for _, c := range countries {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c != "" {
			out[c] = struct{}{}
		}
	}
	return out
}

func productNameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	last := parts[len(parts)-1]
	last = strings.ReplaceAll(last, "-", " ")
	last = strings.ReplaceAll(last, "_", " ")
	return strings.TrimSpace(last)
}

var errEmptyScanQuery = &scanError{msg: "query or productUrl is required"}

type scanError struct{ msg string }

func (e *scanError) Error() string { return e.msg }
