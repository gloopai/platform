package hexmeta

import "testing"

func TestSignMD5_docExample(t *testing.T) {
	params := map[string]string{
		"orderNo":   "9c576c2e-26f2-4bde-96a7-cf14c264e15b",
		"amount":    "10000",
		"name":      "name",
		"phone":     "7277528013",
		"email":     "Djhhkevi@example.xyz",
		"notifyUrl": "",
		"timestamp": "1761120765563",
		"appId":     "023213567912",
	}
	secret := "ycqXEhpZIuZx1JV8yZem9V2I0NA2is0u"
	got := signMD5(params, secret)
	want := "2ef0681c8984dc524d4ef203dacbc31a"
	if got != want {
		t.Fatalf("signMD5: got %q want %q", got, want)
	}
}
