// openapi 开放接口（商户签名 / 收银台 / 回调）的统一错误响应。
package openapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorBody 与成功 JSON 并列，便于商户按 code 分支处理。
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Write(w http.ResponseWriter, httpStatus int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(ErrorBody{Code: code, Message: message})
}

// WriteFromErr 将 gRPC status 或普通 error 映射为 HTTP 状态与业务 code。
func WriteFromErr(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		Write(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpStatus, code, msg := mapGRPC(st.Code(), st.Message())
	Write(w, httpStatus, code, msg)
}

func mapGRPC(c codes.Code, msg string) (httpStatus int, code string, outMsg string) {
	outMsg = msg
	switch c {
	case codes.InvalidArgument:
		return http.StatusBadRequest, "INVALID_ARGUMENT", outMsg
	case codes.NotFound:
		code := "NOT_FOUND"
		if strings.Contains(strings.ToLower(msg), "order") {
			code = "ORDER_NOT_FOUND"
		}
		return http.StatusNotFound, code, outMsg
	case codes.PermissionDenied:
		if strings.Contains(msg, "pay_type not enabled") || strings.Contains(msg, "pay_product_code not enabled") {
			return http.StatusForbidden, "PAY_PRODUCT_NOT_ENABLED", outMsg
		}
		return http.StatusForbidden, "FORBIDDEN", outMsg
	case codes.FailedPrecondition:
		if strings.Contains(strings.ToLower(msg), "no available channel") {
			return http.StatusUnprocessableEntity, "NO_AVAILABLE_CHANNEL", outMsg
		}
		return http.StatusUnprocessableEntity, "FAILED_PRECONDITION", outMsg
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "UNAUTHENTICATED", outMsg
	case codes.Internal:
		return http.StatusInternalServerError, "INTERNAL_ERROR", outMsg
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "UNAVAILABLE", outMsg
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", outMsg
	}
}
