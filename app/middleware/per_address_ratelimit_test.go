package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExtractAddressFromBody_FromAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"from_address":"0xABC","to_address":"0xDEF","amount":10}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(body))

	addr := extractAddressFromBody(c)
	assert.Equal(t, "0xABC", addr)

	// Body harus bisa dibaca ulang oleh handler
	bodyBytes := make([]byte, len(body))
	n, _ := c.Request.Body.Read(bodyBytes)
	assert.Greater(t, n, 0, "body harus bisa dibaca ulang setelah middleware consume")
}

func TestExtractAddressFromBody_Address(t *testing.T) {
	// Untuk buy/sell, field-nya 'address' bukan 'from_address'.
	gin.SetMode(gin.TestMode)
	body := `{"address":"0x123","amount":50}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(body))

	addr := extractAddressFromBody(c)
	assert.Equal(t, "0x123", addr)
}

func TestExtractAddressFromBody_AddressPriority(t *testing.T) {
	// Jika body punya keduanya, 'address' diprioritaskan (buy/sell case).
	gin.SetMode(gin.TestMode)
	body := `{"address":"0xFIRST","from_address":"0xSECOND"}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(body))

	addr := extractAddressFromBody(c)
	assert.Equal(t, "0xFIRST", addr, "'address' diprioritaskan dari 'from_address'")
}

func TestExtractAddressFromBody_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(""))

	addr := extractAddressFromBody(c)
	assert.Equal(t, "", addr)
}

func TestExtractAddressFromBody_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("not json"))

	addr := extractAddressFromBody(c)
	assert.Equal(t, "", addr, "JSON invalid harus return empty, bukan error")
}

func TestExtractAddressFromBody_NoAddressField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"foo":"bar","baz":123}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(body))

	addr := extractAddressFromBody(c)
	assert.Equal(t, "", addr)
}

func TestExtractAddressFromBody_TrimsWhitespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"address":"  0xABC  "}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(body))

	addr := extractAddressFromBody(c)
	assert.Equal(t, "0xABC", addr, "whitespace di-trim")
}
