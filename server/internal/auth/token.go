package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultTokenTTL = 24 * time.Hour

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	err    error
}

type Claims struct {
	UserID    int64
	Username  string
	Role      string
	SessionID string
	ExpiresAt time.Time
}

func NewTokenManagerFromEnv() *TokenManager {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return &TokenManager{err: errors.New("JWT_SECRET is required")}
	}

	ttl := defaultTokenTTL
	if value := strings.TrimSpace(os.Getenv("JWT_EXPIRES_HOURS")); value != "" {
		hours, err := strconv.Atoi(value)
		if err == nil && hours > 0 {
			ttl = time.Duration(hours) * time.Hour
		}
	}
	return NewTokenManager(secret, ttl)
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (m *TokenManager) Sign(userID int64, username, role, sessionID string) (string, int64, error) {
	if err := m.ready(); err != nil {
		return "", 0, err
	}
	if strings.TrimSpace(sessionID) == "" {
		return "", 0, errors.New("session id is required")
	}

	expiresAt := time.Now().Add(m.ttl)
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{
		"sub":      strconv.FormatInt(userID, 10),
		"username": username,
		"role":     role,
		"sid":      sessionID,
		"exp":      expiresAt.Unix(),
	}
	headerPart, err := encodeJWTPart(header)
	if err != nil {
		return "", 0, err
	}
	payloadPart, err := encodeJWTPart(payload)
	if err != nil {
		return "", 0, err
	}
	signingInput := headerPart + "." + payloadPart
	signature := signHS256(signingInput, m.secret)
	return signingInput + "." + signature, int64(m.ttl.Seconds()), nil
}

func (m *TokenManager) Verify(token string) (*Claims, error) {
	if err := m.ready(); err != nil {
		return nil, err
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}
	expected := signHS256(parts[0]+"."+parts[1], m.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errors.New("invalid token signature")
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := decodeJWTPart(parts[0], &header); err != nil {
		return nil, errors.New("invalid token header")
	}
	if header.Alg != "HS256" || header.Typ != "JWT" {
		return nil, errors.New("unsupported token")
	}

	var payload struct {
		Subject  string `json:"sub"`
		Username string `json:"username"`
		Role     string `json:"role"`
		Session  string `json:"sid"`
		Expires  int64  `json:"exp"`
	}
	if err := decodeJWTPart(parts[1], &payload); err != nil {
		return nil, errors.New("invalid token payload")
	}
	userID, err := strconv.ParseInt(payload.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return nil, errors.New("invalid token subject")
	}
	expiresAt := time.Unix(payload.Expires, 0)
	if !time.Now().Before(expiresAt) {
		return nil, errors.New("token expired")
	}
	if payload.Role != RoleStudent && payload.Role != RoleAdmin {
		return nil, errors.New("invalid token role")
	}
	if strings.TrimSpace(payload.Session) == "" {
		return nil, errors.New("invalid token session")
	}
	return &Claims{
		UserID:    userID,
		Username:  payload.Username,
		Role:      payload.Role,
		SessionID: payload.Session,
		ExpiresAt: expiresAt,
	}, nil
}

func (m *TokenManager) ready() error {
	if m == nil {
		return errors.New("token manager is not configured")
	}
	if m.err != nil {
		return m.err
	}
	if len(m.secret) == 0 {
		return errors.New("token secret is not configured")
	}
	return nil
}

func BearerToken(header string) (string, error) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return "", errors.New("authorization header must be Bearer token")
	}
	return fields[1], nil
}

func encodeJWTPart(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeJWTPart(part string, value any) error {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func signHS256(input string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
