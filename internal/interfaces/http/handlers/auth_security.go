package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type tokenClaims struct {
	UserID   uint64 `json:"uid"`
	TenantID uint64 `json:"tid"`
	Exp      int64  `json:"exp"`
}

func hashPassword(password string) (string, error) {
	if password == "" {
		password = "112233"
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return "bcrypt:" + string(hashed), nil
}

func mustHashPassword(password string) string {
	hashed, err := hashPassword(password)
	if err != nil {
		return "bcrypt-error"
	}
	return hashed
}

func verifyPassword(password, stored string) bool {
	if !strings.HasPrefix(stored, "bcrypt:") {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(strings.TrimPrefix(stored, "bcrypt:")), []byte(password)) == nil
}

func issueToken(userID, tenantID uint64, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("token secret is empty")
	}
	claims := tokenClaims{UserID: userID, TenantID: tenantID, Exp: time.Now().Add(ttl).Unix()}
	headerBytes, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	claimBytes, _ := json.Marshal(claims)
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimBytes)
	unsigned := encodedHeader + "." + encodedClaims
	signature := signToken(unsigned, secret)
	return unsigned + "." + signature, nil
}

func parseToken(token, secret string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, errors.New("invalid token")
	}
	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(signToken(unsigned, secret)), []byte(parts[2])) {
		return tokenClaims{}, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, err
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, err
	}
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return tokenClaims{}, errors.New("token expired")
	}
	return claims, nil
}

func signToken(unsigned, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return parts[1]
	}
	return header
}

func parseUintClaim(value string) uint64 {
	parsed, _ := strconv.ParseUint(value, 10, 64)
	return parsed
}
