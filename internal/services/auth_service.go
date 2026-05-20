package services

import (
	"context"
	"time"
)

type AuthService struct {
	cognito  *CognitoAuthService
	legacy   *TokenService
	username string
	password string
}

func NewAuthService(cognito *CognitoAuthService, legacy *TokenService, username, password string) *AuthService {
	return &AuthService{
		cognito:  cognito,
		legacy:   legacy,
		username: username,
		password: password,
	}
}

func (s *AuthService) UsesCognito() bool {
	return s.cognito != nil && s.cognito.Enabled()
}

func (s *AuthService) Login(ctx context.Context, username, password string) (token string, authType string, err error) {
	if s.UsesCognito() {
		idToken, _, err := s.cognito.Authenticate(ctx, username, password)
		if err != nil {
			return "", "", err
		}
		return idToken, "cognito", nil
	}
	if username != s.username || password != s.password {
		return "", "", errInvalidCredentials
	}
	token, err = s.legacy.Generate(username, 12*time.Hour)
	if err != nil {
		return "", "", err
	}
	return token, "legacy", nil
}

func (s *AuthService) Validate(token string) (string, error) {
	if s.UsesCognito() {
		return s.cognito.ValidateToken(token)
	}
	return s.legacy.Validate(token)
}

var errInvalidCredentials = &authError{msg: "invalid credentials"}

type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }
