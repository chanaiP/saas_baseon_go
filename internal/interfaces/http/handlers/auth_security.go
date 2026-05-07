package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
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
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}
	salt := hex.EncodeToString(saltBytes)
	digest := pbkdf2.Key([]byte(password), []byte(salt), 390000, 32, sha256.New)
	return "pbkdf2_sha256$" + salt + "$" + hex.EncodeToString(digest), nil
}

func mustHashPassword(password string) string {
	hashed, err := hashPassword(password)
	if err != nil {
		return "pbkdf2-error"
	}
	return hashed
}

func verifyPassword(password, stored string) bool {
	if !strings.HasPrefix(stored, "pbkdf2_sha256$") {
		return false
	}
	parts := strings.SplitN(stored, "$", 3)
	if len(parts) != 3 {
		return false
	}
	digest := pbkdf2.Key([]byte(password), []byte(parts[1]), 390000, 32, sha256.New)
	return hmac.Equal([]byte(hex.EncodeToString(digest)), []byte(parts[2]))
}

func generateRandomPassword(length int) string {
	n := length
	if n < 8 {
		n = 8
	}
	if n > 128 {
		n = 128
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for {
		var b strings.Builder
		b.Grow(n)
		hasLetter := false
		hasDigit := false
		for i := 0; i < n; i++ {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
			if err != nil {
				return "A1122334455667"
			}
			ch := alphabet[idx.Int64()]
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				hasLetter = true
			}
			if ch >= '0' && ch <= '9' {
				hasDigit = true
			}
			b.WriteByte(ch)
		}
		if hasLetter && hasDigit {
			return b.String()
		}
	}
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
