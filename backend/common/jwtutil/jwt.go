package jwtutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	adminSubjectPrefix    = "admin:"
	merchantSubjectPrefix = "merchant:"
)

func IssueAdminJWT(secret string, adminID int64, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := jwt.RegisteredClaims{
		Subject:   adminSubjectPrefix + strconv.FormatInt(adminID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return s, expiresAt, nil
}

func IssueMerchantJWT(secret, merchantID string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := jwt.RegisteredClaims{
		Subject:   merchantSubjectPrefix + strings.TrimSpace(merchantID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return s, expiresAt, nil
}

func ParseAdminJWT(secret, tokenText string) (int64, error) {
	claims, err := parseClaims(secret, tokenText)
	if err != nil {
		return 0, err
	}
	if !strings.HasPrefix(claims.Subject, adminSubjectPrefix) {
		return 0, fmt.Errorf("invalid token subject")
	}
	idText := strings.TrimPrefix(claims.Subject, adminSubjectPrefix)
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid token subject")
	}
	return id, nil
}

func ParseMerchantJWT(secret, tokenText string) (string, error) {
	claims, err := parseClaims(secret, tokenText)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(claims.Subject, merchantSubjectPrefix) {
		return "", fmt.Errorf("invalid token subject")
	}
	merchantID := strings.TrimSpace(strings.TrimPrefix(claims.Subject, merchantSubjectPrefix))
	if merchantID == "" {
		return "", fmt.Errorf("invalid token subject")
	}
	return merchantID, nil
}

func parseClaims(secret, tokenText string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
