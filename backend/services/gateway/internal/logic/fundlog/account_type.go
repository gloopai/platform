package fundlog

import "strings"

// AccountTypeFromChangeType maps fund_logs.change_type to account bucket for display and filtering.
// payin = 代收余额, available = 可用余额.
func AccountTypeFromChangeType(changeType string) string {
	switch strings.ToUpper(strings.TrimSpace(changeType)) {
	case "PAYOUT_DEBIT":
		return "available"
	default:
		return "payin"
	}
}
