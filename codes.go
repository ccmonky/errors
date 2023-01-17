package errors

import "net/http"

// refer to `https://grpc.github.io/grpc/core/md_doc_statuscodes.html`
var (
	OK                 = NewMetaError(source, "success(0)", "success", status(http.StatusOK))
	Canceled           = NewMetaError(source, "canceled(1)", "client cancelled request", status(499))
	Unknown            = NewMetaError(source, "unknown(2)", "server throws an exception", status(http.StatusInternalServerError))
	InvalidArgument    = NewMetaError(source, "invalid_argument(3)", "invalid argument", status(http.StatusBadRequest))
	DeadlineExceeded   = NewMetaError(source, "deadline_exceeded(4)", "timeout", status(http.StatusGatewayTimeout))
	NotFound           = NewMetaError(source, "not_found(5)", "not found", status(http.StatusNotFound))
	AlreadyExists      = NewMetaError(source, "already_exists(6)", "already exists", status(http.StatusConflict))
	PermissionDenied   = NewMetaError(source, "permission_denied(7)", "permission denied", status(http.StatusForbidden))
	ResourceExhausted  = NewMetaError(source, "resource_exhausted(8)", "resource exhauste", status(http.StatusTooManyRequests))
	FailedPrecondition = NewMetaError(source, "failed_precondition(9)", "failed precondition", status(http.StatusBadRequest))
	Aborted            = NewMetaError(source, "aborted(10)", "operation was aborted", status(http.StatusConflict))
	OutOfRange         = NewMetaError(source, "out_of_range(11)", "operation was attempted past the valid range", status(http.StatusBadRequest))
	Unimplemented      = NewMetaError(source, "unimplemented(12)", "unimplemented", status(http.StatusNotImplemented))
	Internal           = NewMetaError(source, "internal(13)", "internal error", status(http.StatusInternalServerError))
	Unavailable        = NewMetaError(source, "unavailable(14)", "service is unavailable", status(http.StatusServiceUnavailable))
	DataLoss           = NewMetaError(source, "data_loss(15)", "unrecoverable data loss or corruption", status(http.StatusInternalServerError))
	Unauthenticated    = NewMetaError(source, "unauthenticated(16)", "unauthenticated", status(http.StatusForbidden))
)

var (
	source = "github.com/ccmonky/errors"
	status = StatusAttr.Option
)
