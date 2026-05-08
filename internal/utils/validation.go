package utils

import (
	"fmt"
	"regexp"
	"strings"

	"manticore-imports/internal/domain"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func ValidateCreateQuoteRequest(in domain.CreateQuoteRequest) error {
	if strings.TrimSpace(in.CustomerName) == "" {
		return fmt.Errorf("customerName is required")
	}
	if !emailRegex.MatchString(strings.TrimSpace(in.Email)) {
		return fmt.Errorf("valid email is required")
	}
	if strings.TrimSpace(in.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(in.ProductName) == "" {
		return fmt.Errorf("productName is required")
	}
	if strings.TrimSpace(in.ProductCategory) == "" {
		return fmt.Errorf("productCategory is required")
	}
	if in.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if strings.TrimSpace(in.SourceCountry) == "" {
		return fmt.Errorf("sourceCountry is required")
	}
	return nil
}
