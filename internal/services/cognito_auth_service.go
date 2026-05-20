package services

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoAuthService struct {
	client     *cognitoidentityprovider.Client
	userPoolID string
	clientID   string
	region     string
	httpClient *http.Client
	jwksCache  map[string]*rsa.PublicKey
	jwksExpiry time.Time
	jwksMu     sync.RWMutex
}

func NewCognitoAuthService(client *cognitoidentityprovider.Client, userPoolID, clientID, region string) *CognitoAuthService {
	return &CognitoAuthService{
		client:     client,
		userPoolID: userPoolID,
		clientID:   clientID,
		region:     region,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		jwksCache:  make(map[string]*rsa.PublicKey),
	}
}

func (s *CognitoAuthService) Enabled() bool {
	return s != nil && s.userPoolID != "" && s.clientID != ""
}

func (s *CognitoAuthService) Authenticate(ctx context.Context, username, password string) (string, string, error) {
	out, err := s.client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(s.clientID),
		AuthParameters: map[string]string{
			"USERNAME": username,
			"PASSWORD": password,
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("cognito auth failed: %w", err)
	}
	if out.AuthenticationResult == nil {
		return "", "", fmt.Errorf("cognito auth returned no tokens")
	}
	idToken := aws.ToString(out.AuthenticationResult.IdToken)
	accessToken := aws.ToString(out.AuthenticationResult.AccessToken)
	if idToken == "" {
		return "", "", fmt.Errorf("cognito auth returned empty id token")
	}
	return idToken, accessToken, nil
}

func (s *CognitoAuthService) ValidateToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid token header")
	}
	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("invalid token header json")
	}
	if header.Alg != "RS256" {
		return "", fmt.Errorf("unsupported token algorithm")
	}

	pubKey, err := s.getPublicKey(header.Kid)
	if err != nil {
		return "", err
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid token payload")
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", fmt.Errorf("invalid token payload json")
	}

	if err := verifyRS256(parts[0]+"."+parts[1], parts[2], pubKey); err != nil {
		return "", err
	}

	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) < time.Now().Unix() {
		return "", fmt.Errorf("token expired")
	}

	issuer, _ := claims["iss"].(string)
	expectedIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", s.region, s.userPoolID)
	if issuer != expectedIssuer {
		return "", fmt.Errorf("invalid token issuer")
	}

	tokenUse, _ := claims["token_use"].(string)
	if tokenUse != "id" && tokenUse != "access" {
		return "", fmt.Errorf("invalid token use")
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		sub, _ = claims["username"].(string)
	}
	if sub == "" {
		return "", fmt.Errorf("token missing subject")
	}
	return sub, nil
}

func (s *CognitoAuthService) getPublicKey(kid string) (*rsa.PublicKey, error) {
	s.jwksMu.RLock()
	if key, ok := s.jwksCache[kid]; ok && time.Now().Before(s.jwksExpiry) {
		s.jwksMu.RUnlock()
		return key, nil
	}
	s.jwksMu.RUnlock()

	s.jwksMu.Lock()
	defer s.jwksMu.Unlock()

	if key, ok := s.jwksCache[kid]; ok && time.Now().Before(s.jwksExpiry) {
		return key, nil
	}

	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", s.region, s.userPoolID)
	resp, err := s.httpClient.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	cache := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		cache[k.Kid] = pub
	}
	s.jwksCache = cache
	s.jwksExpiry = time.Now().Add(1 * time.Hour)

	key, ok := s.jwksCache[kid]
	if !ok {
		return nil, fmt.Errorf("jwks key not found")
	}
	return key, nil
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}
	var eInt int
	switch len(eBytes) {
	case 0:
		return nil, fmt.Errorf("empty exponent")
	case 1, 2, 3:
		for _, b := range eBytes {
			eInt = eInt<<8 + int(b)
		}
	default:
		eInt = int(binary.BigEndian.Uint32(eBytes[len(eBytes)-4:]))
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: eInt}, nil
}

func verifyRS256(message, signature string, pubKey *rsa.PublicKey) error {
	sig, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid token signature encoding")
	}
	hash := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sig); err != nil {
		return fmt.Errorf("invalid token signature")
	}
	return nil
}
