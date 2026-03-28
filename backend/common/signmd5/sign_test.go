package signmd5

import "testing"

func TestSignSortedKV_deterministic(t *testing.T) {
	sec := "channel_secret_alt"
	p := map[string]string{
		"amount":       "100",
		"event_time":   "1774000000000",
		"merchant_ref": "ORD-1",
		"state":        "SUCCESS",
		"txn_id":       "ALT1",
	}
	a := SignSortedKV(p, sec)
	b := SignSortedKV(p, sec)
	if a != b || len(a) != 32 {
		t.Fatalf("sig=%q len=%d", a, len(a))
	}
}

func TestSignSortedKV_skipsSign(t *testing.T) {
	sec := "s"
	p := map[string]string{"a": "1", "sign": "should_ignore", "b": "2"}
	got := SignSortedKV(p, sec)
	want := SignSortedKV(map[string]string{"a": "1", "b": "2"}, sec)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
