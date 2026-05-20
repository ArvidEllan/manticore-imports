package handlers

import (
	"context"

	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"manticore-imports/internal/domain"
	"manticore-imports/internal/utils"
)

type PublicHandler struct {
	Requests     requestService
	Uploads      uploadService
	DealScanner  dealScannerService
}

func (h *PublicHandler) CreateQuote(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload domain.CreateQuoteRequest
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return utils.Error(http.StatusBadRequest, "invalid json body")
	}
	if err := utils.ValidateCreateQuoteRequest(payload); err != nil {
		return utils.Error(http.StatusBadRequest, err.Error())
	}
	item, err := h.Requests.CreateQuote(context.Background(), payload)
	if err != nil {
		return utils.Error(http.StatusInternalServerError, "failed to create quote request")
	}
	return utils.JSON(http.StatusCreated, item)
}

func (h *PublicHandler) GetStatus(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	reference := req.PathParameters["reference"]
	email := req.QueryStringParameters["email"]
	if reference == "" || email == "" {
		return utils.Error(http.StatusBadRequest, "reference and email are required")
	}
	item, err := h.Requests.LookupByReferenceAndEmail(context.Background(), reference, email)
	if err != nil {
		return utils.Error(http.StatusInternalServerError, "failed to fetch request")
	}
	if item == nil {
		return utils.Error(http.StatusNotFound, "request not found")
	}
	return utils.JSON(http.StatusOK, item)
}

func (h *PublicHandler) CreatePresignedUpload(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload struct {
		RequestID   string `json:"requestId"`
		FileName    string `json:"fileName"`
		ContentType string `json:"contentType"`
	}
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return utils.Error(http.StatusBadRequest, "invalid json body")
	}
	if payload.RequestID == "" || payload.FileName == "" || payload.ContentType == "" {
		return utils.Error(http.StatusBadRequest, "requestId, fileName and contentType are required")
	}
	docID, url, err := h.Uploads.CreatePresignedUpload(context.Background(), payload.RequestID, payload.FileName, payload.ContentType)
	if err != nil {
		return utils.Error(http.StatusInternalServerError, "failed to generate upload url")
	}
	return utils.JSON(http.StatusOK, map[string]string{"documentId": docID, "uploadUrl": url})
}
