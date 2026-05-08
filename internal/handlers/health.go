package handlers

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"manticore-imports/internal/utils"
)

func Health(_ events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return utils.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
