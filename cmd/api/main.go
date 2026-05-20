package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"

	appconfig "manticore-imports/internal/config"
	"manticore-imports/internal/handlers"
	"manticore-imports/internal/repositories"
	"manticore-imports/internal/scanners"
	"manticore-imports/internal/services"
	"manticore-imports/internal/utils"
)

type app struct {
	public *handlers.PublicHandler
	admin  *handlers.AdminHandler
}

func main() {
	ctx := context.Background()
	cfg, err := appconfig.Load()
	if err != nil {
		panic(err)
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		panic(err)
	}

	ddbClient := dynamodb.NewFromConfig(awsCfg)
	s3Client := s3.NewFromConfig(awsCfg)
	sesClient := ses.NewFromConfig(awsCfg)

	requestRepo := repositories.NewRequestRepository(ddbClient, cfg.RequestsTable)
	auditRepo := repositories.NewAuditRepository(ddbClient, cfg.AuditTable)
	emailSvc := services.NewEmailService(sesClient, cfg.SESFromEmail)
	requestSvc := services.NewRequestService(requestRepo, auditRepo, emailSvc)
	uploadSvc := services.NewUploadService(s3Client, cfg.DocumentsBucket)
	tokenSvc := services.NewTokenService(cfg.JWTSecret)
	metricsSvc := services.NewMetricsService(requestRepo)
	dealScannerSvc := services.NewDealScannerService(scanners.DefaultRegistry())

	var cognitoSvc *services.CognitoAuthService
	if cfg.CognitoEnabled() {
		cognitoClient := cognitoidentityprovider.NewFromConfig(awsCfg)
		cognitoSvc = services.NewCognitoAuthService(cognitoClient, cfg.CognitoUserPoolID, cfg.CognitoClientID, cfg.CognitoRegion)
	}
	authSvc := services.NewAuthService(cognitoSvc, tokenSvc, cfg.AdminUsername, cfg.AdminPassword)

	a := &app{
		public: &handlers.PublicHandler{Requests: requestSvc, Uploads: uploadSvc, DealScanner: dealScannerSvc},
		admin: &handlers.AdminHandler{Requests: requestSvc, Auth: authSvc, Metrics: metricsSvc},
	}

	lambda.Start(a.handler)
}

func (a *app) handler(_ context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := strings.ToUpper(req.RequestContext.HTTP.Method)
	path := req.RawPath

	switch {
	case method == http.MethodGet && path == "/health":
		return handlers.Health(req)
	case method == http.MethodPost && path == "/public/quotes":
		return a.public.CreateQuote(req)
	case method == http.MethodGet && strings.HasPrefix(path, "/public/status/"):
		req.PathParameters = mapPathParams(path, "/public/status/", "reference")
		return a.public.GetStatus(req)
	case method == http.MethodPost && path == "/public/uploads/presign":
		return a.public.CreatePresignedUpload(req)
	case method == http.MethodPost && path == "/public/deals/scan":
		return a.public.ScanDeals(req)
	case method == http.MethodPost && path == "/admin/auth/login":
		return a.admin.Login(req)
	case method == http.MethodGet && path == "/admin/requests":
		return a.admin.ListRequests(req)
	case method == http.MethodGet && path == "/admin/metrics":
		return a.admin.GetMetrics(req)
	case method == http.MethodGet && strings.HasPrefix(path, "/admin/requests/") && !strings.HasSuffix(path, "/status"):
		id := strings.TrimPrefix(path, "/admin/requests/")
		if strings.Contains(id, "/") {
			return utils.Error(http.StatusNotFound, "route not found")
		}
		req.PathParameters = map[string]string{"id": id}
		return a.admin.GetRequest(req)
	case method == http.MethodPatch && strings.HasPrefix(path, "/admin/requests/") && strings.HasSuffix(path, "/status"):
		id, err := extractRequestIDFromStatusPath(path)
		if err != nil {
			return utils.Error(http.StatusNotFound, "route not found")
		}
		req.PathParameters = map[string]string{"id": id}
		return a.admin.UpdateStatus(req)
	default:
		return utils.Error(http.StatusNotFound, "route not found")
	}
}

func mapPathParams(path, prefix, name string) map[string]string {
	value := strings.TrimPrefix(path, prefix)
	return map[string]string{name: value}
}

func extractRequestIDFromStatusPath(path string) (string, error) {
	trimmed := strings.TrimPrefix(path, "/admin/requests/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[1] != "status" || parts[0] == "" {
		return "", fmt.Errorf("invalid path")
	}
	return parts[0], nil
}
