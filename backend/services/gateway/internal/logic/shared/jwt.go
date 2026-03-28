package shared

import (
	"time"

	"github.com/gloopai/pay/common/jwtutil"
)

func IssueAdminJWT(secret string, adminID int64, ttl time.Duration) (string, time.Time, error) {
	return jwtutil.IssueAdminJWT(secret, adminID, ttl)
}

func IssueMerchantJWT(secret, merchantID string, ttl time.Duration) (string, time.Time, error) {
	return jwtutil.IssueMerchantJWT(secret, merchantID, ttl)
}

func ParseAdminJWT(secret, tokenText string) (int64, error) {
	return jwtutil.ParseAdminJWT(secret, tokenText)
}

func ParseMerchantJWT(secret, tokenText string) (string, error) {
	return jwtutil.ParseMerchantJWT(secret, tokenText)
}
