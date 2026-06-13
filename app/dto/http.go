package dto

import "github.com/gin-gonic/gin"

// APIResponse adalah envelope universal yang support success DAN error
// response (untuk backward compat dengan 86+ caller yang sudah pakai).
//
// Untuk handler baru, prefer APIErrorResponse (lebih kaya info) atau
// panggil RespondAppError(c, err) yang one-liner.
//
// Field-field:
//   - Success: true untuk sukses, false untuk error
//   - Data: payload sukses (omitempty saat error)
//   - Error: pesan error string (omitempty saat sukses)
//   - Code: HTTP status code - BARU, optional, tidak dipakai legacy callers
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Code    int    `json:"code,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// APIErrorResponse envelope TERSTRUKTUR untuk SEMUA response error.
// Lebih kaya dari APIResponse karena ada error_code + field + details.
//
// Contoh single-field conflict (409):
//
//	{"success":false,"code":409,"error":"address already registered",
//	 "error_code":"ADDRESS_ALREADY_REGISTERED","field":"address"}
//
// Contoh multi-field validation (400):
//
//	{"success":false,"code":400,"error":"validation failed",
//	 "error_code":"VALIDATION_FAILED",
//	 "details":[{"field":"username","message":"field wajib diisi"}]}
type APIErrorResponse struct {
	Success   bool         `json:"success"`
	Code      int          `json:"code"`
	Error     string       `json:"error"`
	ErrorCode ErrorCode    `json:"error_code"`
	Field     string       `json:"field,omitempty"`
	Details   []FieldError `json:"details,omitempty"`
}

// NewSuccessResponse membungkus data ke envelope sukses.
func NewSuccessResponse[T any](data T) APIResponse[T] {
	return APIResponse[T]{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse membungkus error string ke envelope error generic.
// Status code di-set terpisah oleh caller (lihat c.JSON(status, ...) pattern).
//
// DEPRECATED untuk handler baru - prefer RespondAppError(c, err) yang
// auto-extract *AppError dan set status + error_code + field.
// Function ini dipertahankan untuk backward compat dengan 86+ caller
// yang sudah ada.
func NewErrorResponse[T any](errMsg string) APIResponse[T] {
	return APIResponse[T]{
		Success: false,
		Error:   errMsg,
	}
}

// NewAPIErrorResponse membungkus *AppError (atau error generic) ke
// APIErrorResponse. Status code di-set ke appErr.Status (atau 500 fallback).
//
// Return value siap di-pass ke c.JSON(status, resp) atau langsung ke
// c.AbortWithStatusJSON(appErr.Status, dto.NewAPIErrorResponse(err)).
func NewAPIErrorResponse(err error) APIErrorResponse {
	if appErr, ok := AsAppError(err); ok {
		return APIErrorResponse{
			Success:   false,
			Code:      appErr.Status,
			Error:     appErr.Message,
			ErrorCode: appErr.Code,
			Field:     appErr.Field,
		}
	}
	// Fallback: unknown error type. Map ke 500 generic, JANGAN expose
	// err.Error() ke user (bisa bocor SQL/path/internal detail).
	return APIErrorResponse{
		Success:   false,
		Code:      500,
		Error:     "internal server error",
		ErrorCode: CodeInternalError,
	}
}

// RespondAppError adalah one-liner helper untuk handler: extract AppError,
// set status code, dan abort dengan response terstruktur. Pakai ini untuk
// konsistensi di semua handler baru.
//
//	c.BindJSON(...) atau service call -> err
//	if err != nil {
//	    dto.RespondAppError(c, err)
//	    return
//	}
//
// Kalau err bukan *AppError, otomatis di-map ke 500 generic (tidak
// bocor detail). Caller tidak perlu if/else boilerplate.
func RespondAppError(c *gin.Context, err error) {
	resp := NewAPIErrorResponse(err)
	c.AbortWithStatusJSON(resp.Code, resp)
}

// NewValidationErrorResponse khusus untuk multi-field validation error.
// Code HARUS 400 (semua validation = bad request). Returned object bisa
// langsung di-pass ke c.AbortWithStatusJSON(400, resp).
func NewValidationErrorResponse(fields []FieldError) APIErrorResponse {
	return APIErrorResponse{
		Success:   false,
		Code:      400,
		Error:     "validation failed",
		ErrorCode: CodeValidationFailed,
		Details:   fields,
	}
}
