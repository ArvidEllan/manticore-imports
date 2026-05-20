package sites

import (
	"hash/fnv"
	"math"
	"strings"
)

func basePriceUSD(query string) float64 {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return 50
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(normalized))
	seed := h.Sum32()
	// Deterministic price between $15 and $850 based on query hash.
	return 15 + float64(seed%835)
}

func applySiteMultiplier(baseUSD, multiplier float64) float64 {
	return math.Round(baseUSD*multiplier*100) / 100
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

func sellerRating(siteID string, query string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(siteID + query))
	seed := h.Sum32()
	return math.Round((4.0+float64(seed%20)/100)*100) / 100
}
