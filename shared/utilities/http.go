package utilities

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vasapolrittideah/money-tracker-api/shared/contract"
)

// WriteSuccessResponse writes a successful API response with the provided data.
func WriteSuccessResponse(w http.ResponseWriter, r *http.Request, data any, logger *zerolog.Logger) {
	apiResp := &contract.APIResponse{
		Data:      data,
		Timestamp: time.Now(),
	}

	if err := WriteJSON(w, http.StatusOK, apiResp); err != nil {
		logger.Error().Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("failed to write success response")
	}
}

// WriteInternalErrorResponse writes an API error response based on a gRPC error.
func WriteInternalErrorResponse(w http.ResponseWriter, r *http.Request, grpcError error, logger *zerolog.Logger) {
	logger.Error().Err(grpcError).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("internal error occurred")

	st := status.Convert(grpcError)
	errorCode := errorCodeFromGRPCCode(st.Code())
	httpStatus := httpStatusFromGRPCCode(st.Code())

	apiResp := &contract.APIResponse{
		Error: &contract.APIError{
			Code:    errorCode,
			Message: st.Message(),
		},
		Timestamp: time.Now(),
	}

	if err := WriteJSON(w, httpStatus, apiResp); err != nil {
		logger.Error().Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("failed to write error response")
	}
}

// WriteRequestErrorResponse writes a bad request error response with the provided message.
func WriteRequestErrorResponse(w http.ResponseWriter, r *http.Request, message string, logger *zerolog.Logger) {
	logger.Error().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("request error occurred")

	apiResp := &contract.APIResponse{
		Error: &contract.APIError{
			Code:    contract.ErrorCodeBadRequest,
			Message: message,
		},
		Timestamp: time.Now(),
	}

	if err := WriteJSON(w, http.StatusBadRequest, apiResp); err != nil {
		logger.Error().Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("failed to write error response")
	}
}

// WriteValidationErrorResponse writes a validation error response with the provided details.
func WriteValidationErrorResponse(
	w http.ResponseWriter,
	r *http.Request,
	details []contract.APIValidationError,
	logger *zerolog.Logger,
) {
	logger.Error().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("validation error occurred")

	apiResp := &contract.APIResponse{
		Error: &contract.APIError{
			Code:    contract.ErrorCodeValidation,
			Message: "Validation error",
			Details: details,
		},
		Timestamp: time.Now(),
	}

	if err := WriteJSON(w, http.StatusBadRequest, apiResp); err != nil {
		logger.Error().Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("failed to write error response")
	}
}

// errorCodeFromGRPCCode maps gRPC codes to application-specific error codes.
func errorCodeFromGRPCCode(code codes.Code) string {
	switch code {
	case codes.OK:
		return ""
	case codes.Canceled:
		return contract.ErrorCodeInternal
	case codes.Unknown:
		return contract.ErrorCodeInternal
	case codes.InvalidArgument:
		return contract.ErrorCodeBadRequest
	case codes.DeadlineExceeded:
		return contract.ErrorCodeInternal
	case codes.NotFound:
		return contract.ErrorCodeNotFound
	case codes.AlreadyExists:
		return contract.ErrorCodeConflict
	case codes.PermissionDenied:
		return contract.ErrorCodeForbidden
	case codes.ResourceExhausted:
		return contract.ErrorCodeRateLimit
	case codes.Unauthenticated:
		return contract.ErrorCodeUnauthorized
	default:
		return contract.ErrorCodeInternal
	}
}

// httpStatusFromGRPCCode maps gRPC codes to HTTP status codes.
func httpStatusFromGRPCCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
