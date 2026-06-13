package dto

import (
	"errors"
	"net/http"
)

// ErrorCode adalah identifier machine-readable untuk client agar bisa
// switch/case di frontend tanpa harus parse string pesan (yang
// localized-prone). Konvensi: UPPER_SNAKE_CASE.
//
// Tambahkan code baru di sini, jangan reuse untuk kategori berbeda -
// backward compatibility untuk client yang sudah switch by code.
type ErrorCode string

const (
	// 4xx - Client errors. Pesan aman untuk di-expose ke user.
	CodeValidationFailed       ErrorCode = "VALIDATION_FAILED"
	CodeInvalidInput           ErrorCode = "INVALID_INPUT"
	CodeInvalidJSON            ErrorCode = "INVALID_JSON_BODY"
	CodeAddressExists          ErrorCode = "ADDRESS_ALREADY_REGISTERED"
	CodeUsernameExists         ErrorCode = "USERNAME_ALREADY_EXISTS"
	CodeMissingChallenge       ErrorCode = "MISSING_CHALLENGE"
	CodeStaleChallenge         ErrorCode = "STALE_CHALLENGE"
	CodeInvalidSignature       ErrorCode = "INVALID_SIGNATURE"
	CodeSignatureRecoveryFail  ErrorCode = "SIGNATURE_RECOVERY_FAILED"
	CodeAddressMismatch        ErrorCode = "ADDRESS_MISMATCH"
	CodeUserNotFound           ErrorCode = "USER_NOT_FOUND"
	CodeUsernameMismatch       ErrorCode = "USERNAME_MISMATCH"
	CodeUnauthorized           ErrorCode = "UNAUTHORIZED"
	CodeForbidden              ErrorCode = "FORBIDDEN"
	CodeTokenInvalid           ErrorCode = "TOKEN_INVALID"
	CodeTokenExpired           ErrorCode = "TOKEN_EXPIRED"
	CodeTokenRevoked           ErrorCode = "TOKEN_REVOKED"
	CodeNotFound               ErrorCode = "NOT_FOUND"
	CodeConflict               ErrorCode = "CONFLICT"
	CodeRateLimited            ErrorCode = "RATE_LIMITED"
	CodePayloadTooLarge        ErrorCode = "PAYLOAD_TOO_LARGE"

	// 5xx - Server errors. Pesan GENERIC, detail asli di-log internal saja.
	CodeInternalError          ErrorCode = "INTERNAL_ERROR"
	CodeDatabaseError          ErrorCode = "DATABASE_ERROR"
	CodeServiceUnavailable     ErrorCode = "SERVICE_UNAVAILABLE"
	CodeUpstreamError          ErrorCode = "UPSTREAM_ERROR"
)

// AppError adalah error terstruktur yang携带 HTTP status + error code
// + user-facing message + field name (untuk input validation).
//
// Dipakai oleh service layer untuk menandai kategori error secara
// eksplisit, lalu handler layer memetakan ke HTTP response tanpa harus
// string-match err.Error() (yang fragile terhadap i18n/refactor).
//
// Convention:
//   - Status: HTTP status code yang sesuai (400, 409, 500, dll)
//   - Code: machine-readable identifier (lihat const di atas)
//   - Message: human-friendly, AMAN ditampilkan ke user untuk 4xx.
//              Untuk 5xx, JANGAN bocorkan detail internal - log saja.
//   - Field: nama field penyebab (untuk input validation, opsional)
//   - Cause: error original untuk logging/debugging (tidak di-expose)
type AppError struct {
	Status  int
	Code    ErrorCode
	Message string
	Field   string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return string(e.Code) + ": " + e.Message + " (cause: " + e.Cause.Error() + ")"
	}
	return string(e.Code) + ": " + e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

// Is memungkinkan errors.Is(err, target) membandingkan by Code.
// Berguna untuk testing: errors.Is(err, &AppError{Code: CodeAddressExists}).
func (e *AppError) Is(target error) bool {
	var t *AppError
	if !errors.As(target, &t) {
		return false
	}
	return e.Code == t.Code
}

// Constructors. Pakai ini, jangan construct AppError langsung,
// agar tidak ada inconsistent default (mis. lupa set Status).

// NewBadRequest membuat AppError 400 untuk input user yang invalid
// (bukan field-level validation - untuk itu pakai FieldError di details).
func NewBadRequest(code ErrorCode, message string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: message}
}

// NewBadRequestField sama dengan NewBadRequest tapi携带 field name.
// Frontend bisa highlight field yang salah via response.field.
func NewBadRequestField(code ErrorCode, message, field string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: message, Field: field}
}

// NewConflict untuk resource yang sudah ada (duplicate). 409.
func NewConflict(code ErrorCode, message, field string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: code, Message: message, Field: field}
}

// NewNotFound untuk resource yang tidak ditemukan. 404.
func NewNotFound(code ErrorCode, message string) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: code, Message: message}
}

// NewUnauthorized untuk token hilang/invalid. 401.
func NewUnauthorized(code ErrorCode, message string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: code, Message: message}
}

// NewForbidden untuk user yang terautentikasi tapi tidak punya akses. 403.
func NewForbidden(code ErrorCode, message string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: code, Message: message}
}

// NewInternal untuk 5xx. Cause di-log tapi TIDAK di-expose ke client
// (message hard-coded generic, bukan err.Error()).
//
// Pakai pattern:
//
//	if err != nil {
//	    return dto.NewInternalError(err)  // err di-log, tidak bocor
//	}
func NewInternalError(cause error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternalError,
		Message: "internal server error",
		Cause:   cause,
	}
}

// NewDatabaseError alias untuk NewInternalError dengan code DB.
// Tetap generic di message (tidak bocor query/SQL error ke user).
func NewDatabaseError(cause error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    CodeDatabaseError,
		Message: "database error",
		Cause:   cause,
	}
}

// AsAppError ekstrak *AppError dari error chain. Return nil + false
// kalau error bukan AppError (handler fallback ke 500 generic).
func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
