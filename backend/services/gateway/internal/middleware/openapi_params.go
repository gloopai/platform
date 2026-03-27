package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gloopai/pay/gateway/internal/apiresp"
)

type openAPIParamsKeyType struct{}

var openAPIParamsContextKey = openAPIParamsKeyType{}

const defaultMaxOpenAPIBodyBytes int64 = 262144

var errOpenAPIBodyTooLarge = errors.New("request body too large")

// OpenAPIParamsParseMiddleware reads query + JSON body once (with size limit) and stores merged params in request context.
type OpenAPIParamsParseMiddleware struct {
	maxBodyBytes int64
}

func NewOpenAPIParamsParseMiddleware(maxBodyBytes int64) *OpenAPIParamsParseMiddleware {
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxOpenAPIBodyBytes
	}
	return &OpenAPIParamsParseMiddleware{maxBodyBytes: maxBodyBytes}
}

func (m *OpenAPIParamsParseMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := parseParamsFromRequestWithLimit(r, m.maxBodyBytes)
		if err != nil {
			if errors.Is(err, errOpenAPIBodyTooLarge) {
				apiresp.Fail(w, apiresp.CodePayloadTooLarge, "request body too large")
				return
			}
			apiresp.Fail(w, apiresp.CodeInvalidParams, "invalid params")
			return
		}
		ctx := context.WithValue(r.Context(), openAPIParamsContextKey, params)
		next(w, r.WithContext(ctx))
	}
}

func readParams(r *http.Request) (map[string]string, error) {
	if p, ok := r.Context().Value(openAPIParamsContextKey).(map[string]string); ok && p != nil {
		return p, nil
	}
	return parseParamsFromRequestWithLimit(r, defaultMaxOpenAPIBodyBytes)
}

func parseParamsFromRequestWithLimit(r *http.Request, maxBodyBytes int64) (map[string]string, error) {
	params := map[string]string{}
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			params[strings.ToLower(k)] = vs[0]
		}
	}

	ct := r.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		return params, nil
	}

	limit := maxBodyBytes
	if limit <= 0 {
		limit = defaultMaxOpenAPIBodyBytes
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, errOpenAPIBodyTooLarge
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var raw map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, err
		}
	}
	for k, v := range raw {
		if v == nil {
			continue
		}
		params[strings.ToLower(k)] = anyToString(v)
	}
	return params, nil
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
