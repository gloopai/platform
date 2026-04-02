package gatewaymw

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type openAPIParamsKeyType struct{}

var openAPIParamsContextKey = openAPIParamsKeyType{}

// OpenAPIParamsFromContext returns merged query+JSON params set by [OpenAPIParamsParse].
func OpenAPIParamsFromContext(ctx context.Context) (map[string]string, bool) {
	p, ok := ctx.Value(openAPIParamsContextKey).(map[string]string)
	if !ok || p == nil {
		return nil, false
	}
	return p, true
}

const defaultMaxOpenAPIBodyBytes int64 = 262144

var errOpenAPIBodyTooLarge = errors.New("request body too large")

// OpenAPIParamsParseOptions configures [OpenAPIParamsParse].
type OpenAPIParamsParseOptions struct {
	MaxBodyBytes int64
	// Fail writes a JSON error envelope (e.g. apiresp.Fail).
	Fail func(w http.ResponseWriter, code int, message string)
	CodePayloadTooLarge int
	CodeInvalidParams   int
}

// OpenAPIParamsParse reads query + JSON body once (with size limit) and stores merged params in request context.
type OpenAPIParamsParse struct {
	maxBodyBytes        int64
	fail                func(w http.ResponseWriter, code int, message string)
	codePayloadTooLarge int
	codeInvalidParams   int
}

// NewOpenAPIParamsParse builds OpenAPI param middleware. MaxBodyBytes defaults when <= 0.
func NewOpenAPIParamsParse(opt OpenAPIParamsParseOptions) *OpenAPIParamsParse {
	max := opt.MaxBodyBytes
	if max <= 0 {
		max = defaultMaxOpenAPIBodyBytes
	}
	return &OpenAPIParamsParse{
		maxBodyBytes:        max,
		fail:                opt.Fail,
		codePayloadTooLarge: opt.CodePayloadTooLarge,
		codeInvalidParams:   opt.CodeInvalidParams,
	}
}

func (m *OpenAPIParamsParse) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := parseParamsFromRequestWithLimit(r, m.maxBodyBytes)
		if err != nil {
			if errors.Is(err, errOpenAPIBodyTooLarge) {
				m.fail(w, m.codePayloadTooLarge, "request body too large")
				return
			}
			m.fail(w, m.codeInvalidParams, "invalid params")
			return
		}
		ctx := context.WithValue(r.Context(), openAPIParamsContextKey, params)
		next(w, r.WithContext(ctx))
	}
}

// ReadMergedParams returns merged query+JSON params (from context if present, else reads body).
// Used by rate limiting and merchant signing after [OpenAPIParamsParse] may have run.
func ReadMergedParams(r *http.Request) (map[string]string, error) {
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
