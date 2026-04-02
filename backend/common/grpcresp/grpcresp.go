// Package grpcresp maps gRPC errors to HTTP JSON failure envelopes via injectable callbacks.
// Product gateways supply Fail (e.g. gatewayapiresp.Fail) and a domain-specific gRPC→business-code map.
package grpcresp

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Fail writes a business failure envelope (typically HTTP 200 + JSON code).
type Fail func(w http.ResponseWriter, bizCode int, msg string)

// MapGRPC maps gRPC code and message to business code and outbound message.
// For errors that are not *status.Status, c is codes.Unknown and msg is err.Error().
type MapGRPC func(c codes.Code, msg string) (bizCode int, outMsg string)

// WriteFromGRPC writes a failure envelope from err, or no-ops if err is nil.
func WriteFromGRPC(w http.ResponseWriter, err error, fail Fail, mapGRPC MapGRPC) {
	if err == nil {
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		code, msg := mapGRPC(codes.Unknown, err.Error())
		fail(w, code, msg)
		return
	}
	code, msg := mapGRPC(st.Code(), st.Message())
	fail(w, code, msg)
}
