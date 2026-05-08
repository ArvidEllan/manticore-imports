package handlers

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHealth(t *testing.T) {
	resp, err := Health(events.APIGatewayV2HTTPRequest{})
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Body != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %s", resp.Body)
	}
}
