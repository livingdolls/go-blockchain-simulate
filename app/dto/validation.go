package dto

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ethAddressRegex memvalidasi format address Ethereum: 0x + 40 hex char.
var ethAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

// IsEthereumAddress mengecek apakah string adalah address Ethereum valid.
// Trim whitespace sebelum validasi. Tidak ada normalisasi case di sini;
// caller yang menentukan apakah address harus lowercased sebelum dipakai.
func IsEthereumAddress(addr string) bool {
	return ethAddressRegex.MatchString(strings.TrimSpace(addr))
}

// hexStringRegex memvalidasi string hex, dengan atau tanpa prefix `0x`/`0X`.
//
// Pakai regex ini bukan `encoding/hex` Validate karena:
//   - Lebih murah (regex O(n) vs hex.DecodeString yang alokasi []byte)
//   - Error message jelas untuk user (vs error generic dari hex package)
//   - Support `0x` prefix standar Ethereum (alamat, public key, signature)
//
// Format yang diterima:
//   - `abc123`         (plain hex)
//   - `0xabc123`       (dengan 0x prefix)
//   - `0Xabc123`       (uppercase X juga oke)
//
// Yang DITOLAK:
//   - `0x`             (prefix tanpa content - tidak valid hex)
//   - `0xZZ`           (karakter non-hex)
//   - `xyz`            (non-hex tanpa prefix)
//
// Tag ini didaftarkan manual untuk `hex` karena validator/v10 tidak
// punya alias bawaan untuk `hex` (cuma `hexadecimal`).
var hexStringRegex = regexp.MustCompile(`^(0[xX])?[0-9a-fA-F]+$`)

// RegisterCustomValidators mendaftarkan custom validator tags ke validator/v10.
// Panggil SEKALI di main setelah validator/v10 init (biasanya otomatis
// oleh Gin's ShouldBindJSON, tapi tag custom harus di-register eksplisit).
//
// Tag yang tersedia:
//   - eth_addr: validasi Ethereum address (0x + 40 hex char)
//   - hex:      validasi string hex (0-9, a-f, A-F). Bawaan v10 hanya
//     punya `hexadecimal`; `hex` didaftarkan manual sebagai tag pendek.
//
// Pakai di struct: `binding:"required,eth_addr"`.
func RegisterCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("eth_addr", func(fl validator.FieldLevel) bool {
		return IsEthereumAddress(fl.Field().String())
	})
	_ = v.RegisterValidation("hex", func(fl validator.FieldLevel) bool {
		return hexStringRegex.MatchString(fl.Field().String())
	})
}

// FieldError merepresentasikan satu field validation error dengan field
// path dan pesan yang user-friendly. Dipakai untuk response 400.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError agregat beberapa FieldError. Implements error.
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Errors))
	for _, fe := range e.Errors {
		parts = append(parts, fe.Field+": "+fe.Message)
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// AsFieldErrors mengkonversi error dari validator/v10 ke slice of
// FieldError. Field path pakai dot notation (mis. "Address"), cocok
// untuk ditampilkan di frontend.
func AsFieldErrors(err error) []FieldError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []FieldError{{Field: "_", Message: err.Error()}}
	}
	out := make([]FieldError, 0, len(ve))
	for _, fe := range ve {
		out = append(out, FieldError{
			Field:   fieldJSONName(fe),
			Message: validationMessage(fe),
		})
	}
	return out
}

// fieldJSONName mengembalikan nama field JSON (bukan Go field name).
// Validator/v10 tidak expose json tag, jadi kita ambil dari StructField
// reflect (gin.BindJSON mengisi tag dari binding:"required" atau json:"name").
func fieldJSONName(fe validator.FieldError) string {
	// fe.Field() = Go field name (mis. "FromAddress").
	// Pakai field name langsung; JSON binding name sudah digunakan
	// oleh Gin saat parse body. Untuk akurasi 100%, idealnya kita
	// inspect reflect.StructField.Tag.Get("json"), tapi itu butuh
	// akses ke parent struct. Trade-off: pakai Go field name untuk
	// simplicity. Frontend bisa mapping manual.
	return fe.Field()
}

// validationMessage mengkonversi validator tag ke pesan user-friendly
// dalam Bahasa Indonesia. Hanya mendukung subset tag yang dipakai di
// project ini (extend jika butuh).
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field wajib diisi"
	case "min":
		return fmt.Sprintf("nilai minimum %s", fe.Param())
	case "max":
		return fmt.Sprintf("nilai maksimum %s", fe.Param())
	case "gt":
		return fmt.Sprintf("harus lebih besar dari %s", fe.Param())
	case "gte":
		return fmt.Sprintf("harus ≥ %s", fe.Param())
	case "lt":
		return fmt.Sprintf("harus lebih kecil dari %s", fe.Param())
	case "lte":
		return fmt.Sprintf("harus ≤ %s", fe.Param())
	case "len":
		return fmt.Sprintf("panjang harus %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", fe.Param())
	case "eth_addr":
		return "format address Ethereum tidak valid (0x + 40 hex)"
	case "hex":
		return "harus hex string valid"
	case "email":
		return "format email tidak valid"
	default:
		return "tidak valid"
	}
}

// BindJSON helper yang konsisten: parse body, validasi via tag, dan
// return 400 dengan FieldError list jika gagal. Handler cukup panggil
// helper ini tanpa boilerplate error handling.
//
// Response format (lihat NewValidationErrorResponse):
//
//	{
//	  "success": false,
//	  "code": 400,
//	  "error": "validation failed",
//	  "error_code": "VALIDATION_FAILED",
//	  "details": [{"field": "...", "message": "..."}]
//	}
//
// Returns:
//   - true: binding sukses, struct terisi.
//   - false: binding gagal, response 400 sudah di-write. Handler harus return.
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		log.Printf("DEBUG %v", NewValidationErrorResponse(AsFieldErrors(err)))
		c.AbortWithStatusJSON(400, NewValidationErrorResponse(AsFieldErrors(err)))
		return false
	}
	return true
}
