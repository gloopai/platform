package fundlog

import "testing"

func TestAccountTypeFromChangeType(t *testing.T) {
	if got := AccountTypeFromChangeType("PAYOUT_DEBIT"); got != "available" {
		t.Fatalf("PAYOUT_DEBIT: got %q", got)
	}
	if got := AccountTypeFromChangeType("ORDER_PAID"); got != "payin" {
		t.Fatalf("ORDER_PAID: got %q", got)
	}
	if got := AccountTypeFromChangeType("PAYIN_TO_PAYOUT"); got != "payin" {
		t.Fatalf("PAYIN_TO_PAYOUT: got %q", got)
	}
	if got := AccountTypeFromChangeType("AVAILABLE_DEPOSIT"); got != "available" {
		t.Fatalf("AVAILABLE_DEPOSIT: got %q", got)
	}
}
