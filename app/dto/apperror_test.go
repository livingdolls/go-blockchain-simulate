package dto

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Constructors(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		wantStatus int
		wantCode   ErrorCode
		wantField  string
	}{
		{
			name:       "NewBadRequest sets 400",
			err:        NewBadRequest(CodeInvalidInput, "bad input"),
			wantStatus: 400,
			wantCode:   CodeInvalidInput,
		},
		{
			name:       "NewBadRequestField carries field name",
			err:        NewBadRequestField(CodeInvalidInput, "bad input", "email"),
			wantStatus: 400,
			wantCode:   CodeInvalidInput,
			wantField:  "email",
		},
		{
			name:       "NewConflict sets 409",
			err:        NewConflict(CodeAddressExists, "duplicate", "address"),
			wantStatus: 409,
			wantCode:   CodeAddressExists,
			wantField:  "address",
		},
		{
			name:       "NewNotFound sets 404",
			err:        NewNotFound(CodeUserNotFound, "no user"),
			wantStatus: 404,
			wantCode:   CodeUserNotFound,
		},
		{
			name:       "NewUnauthorized sets 401",
			err:        NewUnauthorized(CodeTokenInvalid, "no token"),
			wantStatus: 401,
			wantCode:   CodeTokenInvalid,
		},
		{
			name:       "NewForbidden sets 403",
			err:        NewForbidden(CodeForbidden, "denied"),
			wantStatus: 403,
			wantCode:   CodeForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, tt.err.Status)
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, tt.wantField, tt.err.Field)
		})
	}
}

func TestNewInternalError_HidesCauseMessage(t *testing.T) {
	// CRITICAL: 5xx errors JANGAN bocor detail internal ke user.
	// Message harus generic, cause hanya untuk logging.
	sensitiveErr := errors.New("sql: duplicate entry '0xabc' for key 'users.password_hash'")

	ae := NewInternalError(sensitiveErr)
	assert.Equal(t, 500, ae.Status)
	assert.Equal(t, CodeInternalError, ae.Code)
	assert.Equal(t, "internal server error", ae.Message) // generic, no leak
	assert.NotContains(t, ae.Message, "password_hash", "must not leak sensitive field names")
	assert.NotContains(t, ae.Message, "0xabc", "must not leak data values")

	// Cause tetap accessible untuk logging via Unwrap/errors.Is.
	assert.Same(t, sensitiveErr, ae.Cause)
}

func TestAsAppError_NilSafe(t *testing.T) {
	_, ok := AsAppError(nil)
	assert.False(t, ok, "nil error should return false")
}

func TestAsAppError_ExtractsFromWrap(t *testing.T) {
	inner := NewConflict(CodeAddressExists, "duplicate address", "address")
	// Wrap dengan fmt.Errorf-style
	wrapped := errors.Join(errors.New("wrapped context"), inner)

	ae, ok := AsAppError(wrapped)
	require.True(t, ok)
	assert.Equal(t, CodeAddressExists, ae.Code)
	assert.Equal(t, "address", ae.Field)
}

func TestAppError_Is_ByCode(t *testing.T) {
	// Dua AppError berbeda instance tapi sama code harus match di errors.Is.
	target := &AppError{Code: CodeAddressExists}
	got := NewConflict(CodeAddressExists, "dup", "address")

	assert.True(t, errors.Is(got, target), "should match by Code")
}

func TestNewAPIErrorResponse_FromAppError(t *testing.T) {
	ae := NewConflict(CodeAddressExists, "duplicate", "address")
	resp := NewAPIErrorResponse(ae)

	assert.False(t, resp.Success)
	assert.Equal(t, 409, resp.Code)
	assert.Equal(t, "duplicate", resp.Error)
	assert.Equal(t, CodeAddressExists, resp.ErrorCode)
	assert.Equal(t, "address", resp.Field)
	assert.Empty(t, resp.Details, "single-field error should not have details")
}

func TestNewAPIErrorResponse_FallbackForUnknownError(t *testing.T) {
	// Error yang BUKAN *AppError (mis. raw error dari third-party lib)
	// harus di-map ke 500 generic. JANGAN bocor message.
	sensitive := errors.New("redis: connection refused to internal-host:6379")
	resp := NewAPIErrorResponse(sensitive)

	assert.Equal(t, 500, resp.Code)
	assert.Equal(t, "internal server error", resp.Error)
	assert.NotContains(t, resp.Error, "redis")
	assert.NotContains(t, resp.Error, "internal-host")
}

func TestNewValidationErrorResponse_HasDetails(t *testing.T) {
	details := []FieldError{
		{Field: "username", Message: "wajib diisi"},
		{Field: "address", Message: "format tidak valid"},
	}
	resp := NewValidationErrorResponse(details)

	assert.False(t, resp.Success)
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, CodeValidationFailed, resp.ErrorCode)
	assert.Len(t, resp.Details, 2)
	assert.Equal(t, "username", resp.Details[0].Field)
}

func TestAppError_ErrorString_IncludesCode(t *testing.T) {
	// Error() harus menyertakan code agar log/monitoring bisa grep by code.
	ae := NewBadRequest(CodeInvalidSignature, "bad sig")
	// Minimal code ada di string
	assert.Contains(t, ae.Error(), string(CodeInvalidSignature))

	// Dengan cause
	ae2 := NewInternalError(errors.New("underlying"))
	assert.Contains(t, ae2.Error(), string(CodeInternalError))
	assert.Contains(t, ae2.Error(), "underlying")
}

func TestRespondAppError_Integration(t *testing.T) {
	// End-to-end: pass error ke RespondAppError, verify HTTP response shape.
	// Ini test handler layer (mock Gin context).
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		err := NewConflict(CodeAddressExists, "duplicate", "address")
		RespondAppError(c, err)
	})

	rec := serve(r, newReq("POST", "/test", strings.NewReader("")))
	assert.Equal(t, http.StatusConflict, rec.Code)

	// Body harus match APIErrorResponse shape
	body := rec.Body.String()
	assert.Contains(t, body, `"success":false`)
	assert.Contains(t, body, `"code":409`)
	assert.Contains(t, body, `"error_code":"ADDRESS_ALREADY_REGISTERED"`)
	assert.Contains(t, body, `"field":"address"`)
}
