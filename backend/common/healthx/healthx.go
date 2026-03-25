package healthx

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type CheckFunc func(ctx context.Context) error

type Check struct {
	Name string
	Fn   CheckFunc
}

type Result struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
	LatencyMs int64 `json:"latency_ms,omitempty"`
}

// HTTPHandler returns:
// - 200 when all checks pass
// - 503 when any check fails
func HTTPHandler(checks ...Check) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		results := make([]Result, 0, len(checks))
		allOK := true
		for _, c := range checks {
			start := time.Now()
			err := c.Fn(ctx)
			lat := time.Since(start).Milliseconds()
			res := Result{Name: c.Name, OK: err == nil, LatencyMs: lat}
			if err != nil {
				allOK = false
				res.Error = err.Error()
			}
			results = append(results, res)
		}

		body := map[string]any{
			"ok":       allOK,
			"checks":   results,
			"ts_ms":    time.Now().UnixMilli(),
		}
		w.Header().Set("Content-Type", "application/json")
		if allOK {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(body)
	}
}

