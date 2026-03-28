package mockpsp2

import "testing"

func TestSignMd5SortedKV_deterministic(t *testing.T) {
	sec := "channel_secret_alt"
	p := map[string]string{
		"amount":       "100",
		"event_time":   "1774000000000",
		"merchant_ref": "ORD-1",
		"state":        "SUCCESS",
		"txn_id":       "ALT1",
	}
	a := SignMd5SortedKV(p, sec)
	b := SignMd5SortedKV(p, sec)
	if a != b || len(a) != 32 {
		t.Fatalf("sig=%q len=%d", a, len(a))
	}
}
