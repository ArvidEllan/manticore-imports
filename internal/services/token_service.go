package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TokenService struct {
	secret string
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: secret}
}

func (s *TokenService) Generate(subject string, ttl time.Duration) (string, error) {
	head := map[string]string{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{
		"sub": subject,
		"exp": time.Now().Add(ttl).Unix(),
	}
	return signJWT(head, payload, s.secret)
}

func (s *TokenService) Validate(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token")
	}
	message := parts[0] + "." + parts[1]
	expected := sign(message, s.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", fmt.Errorf("invalid token signature")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid payload")
	}
	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", fmt.Errorf("invalid payload json")
	}
	exp, ok := payload["exp"].(float64)
	if !ok || int64(exp) < time.Now().Unix() {
		return "", fmt.Errorf("token expired")
	}
	sub, _ := payload["sub"].(string)
	return sub, nil
}

func signJWT(header map[string]string, payload map[string]any, secret string) (string, error) {
	hb, err := json.Marshal(header)
	if err != nil { return "", err }
	pb, err := json.Marshal(payload)
	if err != nil { return "", err }
	headEnc := base64.RawURLEncoding.EncodeToString(hb)
	payloadEnc := base64.RawURLEncoding.EncodeToString(pb)
	message := headEnc + "." + payloadEnc
	return message + "." + sign(message, secret), nil
}

func sign(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
