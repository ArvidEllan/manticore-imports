package email

import (
	"strings"
	"testing"
)

func TestRenderQuoteReceived(t *testing.T) {
	html, text, err := RenderQuoteReceived(QuoteReceivedData{
		CustomerName:  "Jane",
		Reference:     "MANT-20260520-ABCDE",
		ProductName:   "Pump",
		Quantity:      2,
		SourceCountry: "CN",
	})
	if err != nil {
		t.Fatalf("RenderQuoteReceived returned error: %v", err)
	}
	if !strings.Contains(html, "Jane") || !strings.Contains(html, "MANT-20260520-ABCDE") {
		t.Fatalf("html missing expected content: %s", html)
	}
	if !strings.Contains(text, "Pump") {
		t.Fatalf("text missing product name: %s", text)
	}
}

func TestRenderStatusUpdated(t *testing.T) {
	html, text, err := RenderStatusUpdated(StatusUpdatedData{
		CustomerName: "Jane",
		Reference:    "MANT-20260520-ABCDE",
		Status:       "UNDER_REVIEW",
	})
	if err != nil {
		t.Fatalf("RenderStatusUpdated returned error: %v", err)
	}
	if !strings.Contains(html, "UNDER_REVIEW") {
		t.Fatalf("html missing status: %s", html)
	}
	if !strings.Contains(text, "UNDER_REVIEW") {
		t.Fatalf("text missing status: %s", text)
	}
}
