package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"manticore-imports/internal/domain"
	"manticore-imports/internal/utils"
)

func (h *PublicHandler) ScanDeals(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if h.DealScanner == nil {
		return utils.Error(http.StatusServiceUnavailable, "deal scanner not configured")
	}
	var payload domain.ScanDealsRequest
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return utils.Error(http.StatusBadRequest, "invalid json body")
	}
	if err := utils.ValidateScanDealsRequest(payload); err != nil {
		return utils.Error(http.StatusBadRequest, err.Error())
	}
	result, err := h.DealScanner.Scan(context.Background(), payload)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			return utils.Error(http.StatusBadRequest, err.Error())
		}
		return utils.Error(http.StatusInternalServerError, "failed to scan deals")
	}
	return utils.JSON(http.StatusOK, result)
}
