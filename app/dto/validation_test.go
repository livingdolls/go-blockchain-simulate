package dto

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsEthereumAddress(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid lowercase", "0x1234567890abcdef1234567890abcdef12345678", true},
		{"valid uppercase", "0xABCDEF1234567890ABCDEF1234567890ABCDEF12", true},
		{"valid mixed case", "0xAbCdEf1234567890aBcDeF1234567890AbCdEf12", true},
		{"empty", "", false},
		{"missing 0x prefix", "1234567890abcdef1234567890abcdef12345678", false},
		{"too short", "0x1234567890abcdef", false},
		{"too long", "0x1234567890abcdef1234567890abcdef1234567890", false},
		{"non-hex char", "0x1234567890abcdef1234567890abcdef1234567g", false},
		{"with whitespace", "  0x1234567890abcdef1234567890abcdef12345678  ", true},
		{"all zeros", "0x0000000000000000000000000000000000000000", true},
		{"all f's", "0xffffffffffffffffffffffffffffffffffffffff", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsEthereumAddress(tt.in))
		})
	}
}

func TestAsFieldErrors_ValidatorError(t *testing.T) {
	// Buat validator error dari struct invalid.
	v := validator.New()
	type testStruct struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=0,lte=150"`
	}
	s := testStruct{Email: "not-an-email", Age: 200}
	err := v.Struct(s)
	require.Error(t, err)

	fieldErrs := AsFieldErrors(err)
	require.NotEmpty(t, fieldErrs)

	// Harus ada 2 field error: Email dan Age
	fields := make(map[string]string)
	for _, fe := range fieldErrs {
		fields[fe.Field] = fe.Message
	}
	assert.Contains(t, fields, "Email")
	assert.Contains(t, fields, "Age")
}

func TestAsFieldErrors_NonValidatorError(t *testing.T) {
	// Plain error (bukan validator.ValidationErrors) -> single entry
	// dengan field "_".
	err := errors.New("something broke")
	fieldErrs := AsFieldErrors(err)
	require.Len(t, fieldErrs, 1)
	assert.Equal(t, "_", fieldErrs[0].Field)
	assert.Equal(t, "something broke", fieldErrs[0].Message)
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{Errors: []FieldError{
		{Field: "Email", Message: "wajib diisi"},
		{Field: "Age", Message: "harus > 0"},
	}}
	msg := ve.Error()
	assert.Contains(t, msg, "Email: wajib diisi")
	assert.Contains(t, msg, "Age: harus > 0")
}

func TestBindJSON_ValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	type req struct {
		Address string `json:"address" binding:"required,eth_addr"`
	}
	r.POST("/test", func(c *gin.Context) {
		var body req
		if !BindJSON(c, &body) {
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})

	// Address invalid
	body := strings.NewReader(`{"address": "not-eth-addr"}`)
	httpReq := newReq("POST", "/test", body)
	rec := serve(r, httpReq)
	assert.Equal(t, 400, rec.Code)
	assert.Contains(t, rec.Body.String(), "validation failed")
	assert.Contains(t, rec.Body.String(), "Address")
}

func TestBindJSON_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	type req struct {
		Address string `json:"address" binding:"required,eth_addr"`
	}
	r.POST("/test", func(c *gin.Context) {
		var body req
		if !BindJSON(c, &body) {
			return
		}
		c.JSON(200, gin.H{"address": body.Address})
	})

	body := strings.NewReader(`{"address": "0x1234567890abcdef1234567890abcdef12345678"}`)
	httpReq := newReq("POST", "/test", body)
	rec := serve(r, httpReq)
	assert.Equal(t, 200, rec.Code)
}

func TestRegisterCustomValidators_EthAddr(t *testing.T) {
	// Verifikasi bahwa RegisterCustomValidators mendaftarkan 'eth_addr'.
	v := validator.New()
	RegisterCustomValidators(v)

	type testStruct struct {
		Address string `validate:"eth_addr"`
	}

	// Invalid
	err := v.Struct(testStruct{Address: "garbage"})
	assert.Error(t, err, "eth_addr harus reject 'garbage'")

	// Valid
	err = v.Struct(testStruct{Address: "0x1234567890abcdef1234567890abcdef12345678"})
	assert.NoError(t, err, "eth_addr harus accept address valid")
}

func TestRegisterCustomValidators_Hex(t *testing.T) {
	// Verifikasi tag 'hex' yang didaftarkan manual. Penting karena:
	//   - Tag ini dipakai untuk public_key (uncompressed ECDSA = 130 hex char)
	//   - Ethereum convention: public key dikirim dengan prefix '0x'
	//   - Bug awal: regex tidak allow '0x' prefix, jadi semua key dengan
	//     prefix di-reject sebagai "harus hex string valid"
	v := validator.New()
	RegisterCustomValidators(v)

	type testStruct struct {
		Key string `validate:"hex"`
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid cases
		{"plain hex lowercase", "abc123def456", false},
		{"plain hex uppercase", "ABC123DEF456", false},
		{"plain hex mixed", "AbC123dEf456", false},
		{"with 0x prefix", "0xabc123def456", false},
		{"with 0X prefix uppercase", "0XABC123DEF456", false},
		{"with 0x prefix mixed", "0xAbC123dEf456", false},
		{"public key size (65 bytes)", "0x0408bcface31732a0d1218cf2bdd8f39113d04ed4984f3338036cdb385daf822bc5b658508a31ffe54d50cbf9dbed62ac7bcef5c595f96af4b3ab56e4475e31124", false},

		// Invalid cases
		{"empty", "", true},
		{"non-hex char", "xyz", true},
		{"0x prefix only (no content)", "0x", true},
		{"0X prefix only (no content)", "0X", true},
		{"non-hex with 0x prefix", "0xZZZ", true},
		{"only 0x + non-hex", "0xGHIJKL", true},
		{"contains space", "abc 123", true},
		{"contains newline", "abc\n123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(testStruct{Key: tt.input})
			if tt.wantErr {
				assert.Error(t, err, "input %q harus di-reject", tt.input)
			} else {
				assert.NoError(t, err, "input %q harus di-accept", tt.input)
			}
		})
	}
}

// helpers untuk test BindJSON
func newReq(method, target string, body *strings.Reader) *http.Request {
	req, _ := http.NewRequest(method, target, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func serve(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}
