// apiresp 网关统一 JSON：{"code":int,"message":string,"data":object}；业务成败以 code 为准，HTTP 状态对业务接口固定 200（探活等例外见各 Handler）。
// 信封写入见 [github.com/gloopai/platform/common/gatewayapiresp]；本包保留平台域业务码与 gRPC 映射。
package apiresp

import (
	"net/http"

	"github.com/gloopai/platform/common/gatewayapiresp"
	"github.com/gloopai/platform/common/grpcresp"
	"google.golang.org/grpc/codes"
)

// CodeSuccess 业务成功（与 gatewayapiresp 一致）。
const CodeSuccess = gatewayapiresp.CodeSuccess

const (
	CodeInvalidParams = 4001

	CodeUnauthorized = 4010

	CodeForbidden = 4030

	CodeNotFound = 4040

	CodePayloadTooLarge = 4130

	CodeTooManyRequests = 4290

	CodeFailedPrecondition = 4220

	CodeInternal    = 5000
	CodeUnavailable = 5003
)

// OK 业务成功：HTTP 200，code=2000，payload 放入 data。
func OK(w http.ResponseWriter, data any) {
	gatewayapiresp.OK(w, data)
}

// Fail 业务失败：HTTP 200，data 为空对象。
func Fail(w http.ResponseWriter, code int, message string) {
	gatewayapiresp.Fail(w, code, message)
}

// FailStatus 用于就绪探针等需保留 HTTP 状态语义的场景；body 仍为统一 envelope。
func FailStatus(w http.ResponseWriter, httpStatus int, code int, message string) {
	gatewayapiresp.FailStatus(w, httpStatus, code, message)
}

// OKStatus 成功但需非 200 的 HTTP 状态（一般不用）。
func OKStatus(w http.ResponseWriter, httpStatus int, data any) {
	gatewayapiresp.OKStatus(w, httpStatus, data)
}

// WriteFromGRPC 将 gRPC 错误映射为业务 code，HTTP 恒为 200。
func WriteFromGRPC(w http.ResponseWriter, err error) {
	grpcresp.WriteFromGRPC(w, err, Fail, grpcToBiz)
}

func grpcToBiz(c codes.Code, msg string) (code int, outMsg string) {
	outMsg = msg
	switch c {
	case codes.InvalidArgument:
		return CodeInvalidParams, outMsg
	case codes.NotFound:
		return CodeNotFound, outMsg
	case codes.PermissionDenied:
		return CodeForbidden, outMsg
	case codes.FailedPrecondition:
		return CodeFailedPrecondition, outMsg
	case codes.Unauthenticated:
		return CodeUnauthorized, outMsg
	case codes.Internal:
		return CodeInternal, outMsg
	case codes.Unavailable:
		return CodeUnavailable, outMsg
	default:
		return CodeInternal, outMsg
	}
}
