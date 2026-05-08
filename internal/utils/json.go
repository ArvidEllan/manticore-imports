package utils

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func JSON(status int, payload any) (events.APIGatewayV2HTTPResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Error(status, "failed to serialize response")
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func Error(status int, message string) (events.APIGatewayV2HTTPResponse, error) {
	return JSON(status, map[string]string{"error": message})
}
