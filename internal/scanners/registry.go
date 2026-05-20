package scanners

import (
	"manticore-imports/internal/scanners/sites"
)

func DefaultRegistry() []Scanner {
	return []Scanner{
		sites.NewAmazon("amazon-us", "Amazon US", "US", "USD", "amazon.com"),
		sites.NewAmazon("amazon-uk", "Amazon UK", "GB", "GBP", "amazon.co.uk"),
		sites.NewAmazon("amazon-de", "Amazon Germany", "DE", "EUR", "amazon.de"),
		sites.NewAlibaba(),
		sites.NewAliExpress(),
		sites.NewEbay("ebay-us", "eBay US", "US", "USD", "ebay.com"),
		sites.NewEbay("ebay-uk", "eBay UK", "GB", "GBP", "ebay.co.uk"),
	}
}
