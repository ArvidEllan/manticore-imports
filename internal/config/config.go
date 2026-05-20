package config

import (
	"fmt"
	"os"
)

type Config struct {
	AWSRegion          string
	Stage              string
	RequestsTable      string
	AuditTable         string
	DocumentsTable     string
	DocumentsBucket    string
	SESFromEmail       string
	AdminUsername      string
	AdminPassword      string
	JWTSecret          string
	CognitoUserPoolID  string
	CognitoClientID    string
	CognitoRegion      string
	FrontendBucket     string
	CloudFrontDomain   string
}

func Load() (Config, error) {
	cfg := Config{
		AWSRegion:         getEnv("AWS_REGION", "eu-west-1"),
		Stage:             getEnv("STAGE", "dev"),
		RequestsTable:     os.Getenv("REQUESTS_TABLE"),
		AuditTable:        os.Getenv("AUDIT_TABLE"),
		DocumentsTable:    os.Getenv("DOCUMENTS_TABLE"),
		DocumentsBucket:   os.Getenv("DOCUMENTS_BUCKET"),
		SESFromEmail:      os.Getenv("SES_FROM_EMAIL"),
		AdminUsername:     os.Getenv("ADMIN_USERNAME"),
		AdminPassword:     os.Getenv("ADMIN_PASSWORD"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		CognitoUserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		CognitoClientID:   os.Getenv("COGNITO_CLIENT_ID"),
		CognitoRegion:     getEnv("COGNITO_REGION", getEnv("AWS_REGION", "eu-west-1")),
		FrontendBucket:    os.Getenv("FRONTEND_BUCKET"),
		CloudFrontDomain:  os.Getenv("CLOUDFRONT_DOMAIN"),
	}

	if cfg.RequestsTable == "" || cfg.AuditTable == "" || cfg.DocumentsTable == "" || cfg.DocumentsBucket == "" {
		return Config{}, fmt.Errorf("missing required table or bucket configuration")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (c Config) CognitoEnabled() bool {
	return c.CognitoUserPoolID != "" && c.CognitoClientID != ""
}
