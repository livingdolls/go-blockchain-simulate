package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySizeMiddleware membatasi ukuran body request untuk mencegah
// DoS via JSON bomb (giant payload) atau file upload tak terbatas.
//
// Default 1 MiB cukup untuk seluruh endpoint JSON aplikasi ini
// (transaction, balance, block). Untuk endpoint yang butuh upload
// file besar, pasang middleware ini di group yang lebih spesifik
// dengan limit lebih besar.
//
// Limit di-check SEBELUM handler dipanggil, sehingga handler tidak
// menerima body yang sudah kebablasan ukuran.
func MaxBodySizeMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// MaxBytesReader returns error jika body > maxBytes.
		// Handler yang pakai c.ShouldBindJSON akan dapat error
		// "http: request body too large".
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
