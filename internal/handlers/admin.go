package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"manticore-imports/internal/domain"
	"manticore-imports/internal/utils"
)

type AdminHandler struct {
	Requests     requestService
	Auth         authService
	Metrics      metricsService
}

func (h *AdminHandler) Login(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return utils.Error(http.StatusBadRequest, "invalid json body")
	}
	token, authType, err := h.Auth.Login(context.Background(), payload.Username, payload.Password)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") || strings.Contains(err.Error(), "cognito auth failed") {
			return utils.Error(http.StatusUnauthorized, "invalid credentials")
		}
		return utils.Error(http.StatusInternalServerError, "failed to issue token")
	}
	return utils.JSON(http.StatusOK, map[string]string{
		"token":    token,
		"authType": authType,
	})
}

func (h *AdminHandler) ListRequests(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if err := h.authorize(req); err != nil {
		return utils.Error(http.StatusUnauthorized, err.Error())
	}
	limit := int32(25)
	if raw := req.QueryStringParameters["limit"]; raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 32); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}
	page, err := h.Requests.ListPage(context.Background(), domain.ListRequestsParams{
		Limit:  limit,
		Cursor: req.QueryStringParameters["cursor"],
	})
	if err != nil {
		return utils.Error(http.StatusInternalServerError, "failed to list requests")
	}
	return utils.JSON(http.StatusOK, page)
}

func (h *AdminHandler) GetRequest(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if err := h.authorize(req); err != nil {
		return utils.Error(http.StatusUnauthorized, err.Error())
	}
	item, err := h.Requests.GetByID(context.Background(), req.PathParameters["id"])
	if err != nil {
		return utils.Error(http.StatusInternalServerError, "failed to fetch request")
	}
	if item == nil {
		return utils.Error(http.StatusNotFound, "request not found")
	}
	return utils.JSON(http.StatusOK, item)
}

func (h *AdminHandler) UpdateStatus(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	subject, err := h.authorizeWithSubject(req)
	if err != nil {
		return utils.Error(http.StatusUnauthorized, err.Error())
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return utils.Error(http.StatusBadRequest, "invalid json body")
	}
	if _, ok := domain.AllowedStatuses[payload.Status]; !ok {
		return utils.Error(http.StatusBadRequest, "invalid status")
	}
	if err := h.Requests.UpdateStatus(context.Background(), req.PathParameters["id"], payload.Status, subject); err != nil {
		return utils.Error(http.StatusInternalServerError, "failed to update status")
	}
	return utils.JSON(http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *AdminHandler) GetMetrics(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if err := h.authorize(req); err != nil {
		return utils.Error(http.StatusUnauthorized, err.Error())
	}
	if h.Metrics == nil {
		return utils.Error(http.StatusServiceUnavailable, "metrics not configured")
	}
	snapshot, err := h.Metrics.Snapshot(context.Background())
	if err != nil {
		return utils.Error(http.StatusInternalServerError, "failed to fetch metrics")
	}
	return utils.JSON(http.StatusOK, snapshot)
}

func (h *AdminHandler) authorize(req events.APIGatewayV2HTTPRequest) error {
	_, err := h.authorizeWithSubject(req)
	return err
}

func (h *AdminHandler) authorizeWithSubject(req events.APIGatewayV2HTTPRequest) (string, error) {
	header := req.Headers["authorization"]
	if header == "" {
		header = req.Headers["Authorization"]
	}
	if !strings.HasPrefix(header, "Bearer ") {
		return "", http.ErrNoCookie
	}
	return h.Auth.Validate(strings.TrimPrefix(header, "Bearer "))
}
