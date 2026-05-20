package domain

import "time"

type ProductListing struct {
	SiteID           string  `json:"siteId"`
	SiteName         string  `json:"siteName"`
	Country          string  `json:"country"`
	Currency         string  `json:"currency"`
	Price            float64 `json:"price"`
	PriceUSD         float64 `json:"priceUsd"`
	ShippingEstimate float64 `json:"shippingEstimateUsd"`
	TotalUSD         float64 `json:"totalUsd"`
	ProductName      string  `json:"productName"`
	ProductURL       string  `json:"productUrl"`
	InStock          bool    `json:"inStock"`
	SellerRating     float64 `json:"sellerRating"`
}

type ScanDealsRequest struct {
	Query      string   `json:"query"`
	ProductURL string   `json:"productUrl"`
	Countries  []string `json:"countries"`
	MaxResults int      `json:"maxResults"`
}

type SiteScanError struct {
	SiteID string `json:"siteId"`
	Error  string `json:"error"`
}

type ScanDealsResult struct {
	ScanID       string           `json:"scanId"`
	Query        string           `json:"query"`
	BestDeal     *ProductListing  `json:"bestDeal,omitempty"`
	Listings     []ProductListing `json:"listings"`
	ScannedSites []string         `json:"scannedSites"`
	FailedSites  []SiteScanError  `json:"failedSites,omitempty"`
	ScannedAt    time.Time        `json:"scannedAt"`
}
