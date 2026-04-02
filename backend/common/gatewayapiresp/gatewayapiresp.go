// Package gatewayapiresp provides the shared JSON envelope for product gateways:
// {"code":int,"message":string,"data":object}. Business outcome is determined by code; HTTP is usually 200.
package gatewayapiresp

import (
	"encoding/json"
	"net/http"
)

// CodeSuccess is the standard business success code (HTTP 200, code 2000).
const CodeSuccess = 2000

type envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

var emptyObj = map[string]any{}

// Write writes the unified envelope with the given HTTP status and business code.
func Write(w http.ResponseWriter, httpStatus int, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	if data == nil {
		data = emptyObj
	}
	_ = json.NewEncoder(w).Encode(envelope{Code: code, Message: message, Data: data})
}

// OK responds with HTTP 200, CodeSuccess, and payload in data.
func OK(w http.ResponseWriter, data any) {
	Write(w, http.StatusOK, CodeSuccess, "", data)
}

// Fail responds with HTTP 200, the given business code, and empty data object.
func Fail(w http.ResponseWriter, code int, message string) {
	Write(w, http.StatusOK, code, message, emptyObj)
}

// FailStatus is for probes etc. where HTTP status must be preserved; body uses the same envelope.
func FailStatus(w http.ResponseWriter, httpStatus int, code int, message string) {
	Write(w, httpStatus, code, message, emptyObj)
}

// OKStatus succeeds with a non-200 HTTP status (rare).
func OKStatus(w http.ResponseWriter, httpStatus int, data any) {
	Write(w, httpStatus, CodeSuccess, "", data)
}
