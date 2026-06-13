package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

// setAuthCookie menulis JWT ke cookie browser dengan SameSite=Strict.
// Set-Cookie via http.SetCookie langsung agar bisa specify SameSite
// (gin.Context.SetCookie tidak support parameter ini).
//
// maxAgeSeconds adalah lifetime cookie dalam detik. Gunakan nilai negatif
// untuk delete cookie (logout).
//
// SameSite=Strict dipilih (bukan Lax) karena auth_token tidak perlu
// dikirim untuk cross-site GET; semua endpoint yang butuh auth adalah
// state-changing (POST/PUT/DELETE). Trade-off: link eksternal ke
// dashboard akan butuh login ulang setelah navigate, tapi ini
// acceptable untuk aplikasi ini.
//
// Secure flag di-skip (false) untuk kompatibilitas development HTTP.
// Production harus override via reverse proxy (HTTPS) atau tambah
// middleware force-redirect ke HTTPS.
func setAuthCookie(c *gin.Context, name, token string, maxAgeSeconds int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAgeSeconds,
		HttpOnly: true,
		Secure:   false, // dev mode; production via reverse proxy HTTPS
		SameSite: http.SameSiteStrictMode,
	})
}

// blacklistToken menghitung hash token dan menyimpannya di Redis
// blacklist dengan TTL. Dipanggil saat logout untuk mencegah token
// yang sudah di-logout dipakai ulang.
func blacklistToken(c *gin.Context, token string, memory redis.MemoryAdapter) {
	if memory == nil || strings.TrimSpace(token) == "" {
		return
	}
	hash := security.TokenHash(token)
	// TTL 24 jam (sama dengan max token lifetime). Redis otomatis
	// menghapus entry setelah TTL habis, sehingga tidak perlu cleanup manual.
	memory.Set(c.Request.Context(), blacklistPrefix+hash, []byte("1"), 24*3600)
}

type RegisterHandler struct {
	service services.RegisterService
	memory  redis.MemoryAdapter
}

func NewRegisterHandler(service services.RegisterService, memory redis.MemoryAdapter) *RegisterHandler {
	return &RegisterHandler{service: service, memory: memory}
}

func (h *RegisterHandler) Register(c *gin.Context) {
	var req models.UserRegister
	if !dto.BindJSON(c, &req) {
		return
	}

	fmt.Errorf("error disini")

	user, err := h.service.Register(req)
	if err != nil {
		// RespondAppError otomatis extract *dto.AppError dari error chain,
		// set HTTP status code yang sesuai (400/409/500), dan return
		// response terstruktur dengan error_code + field. Kalau err
		// bukan AppError, fallback ke 500 generic (tidak bocor detail).
		dto.RespondAppError(c, err)
		return
	}

	resp := &models.UserRegisterResponse{
		Address:  user.Address,
		Username: user.Username,
	}

	setAuthCookie(c, "auth_token", user.Token, 24*3600)

	c.JSON(http.StatusCreated, dto.NewSuccessResponse(resp))
}

func (h *RegisterHandler) Challenge(c *gin.Context) {
	address := c.Param("address")

	// Validasi format address sebelum generate nonce. Tanpa validasi,
	// attacker bisa isi Redis dengan non-address keys (setiap key punya
	// TTL 10 menit → Redis memory exhaustion vector).
	if !dto.IsEthereumAddress(address) {
		dto.RespondAppError(c, dto.NewBadRequestField(dto.CodeInvalidInput,
			"invalid Ethereum address format (expected 0x + 40 hex)", "address"))
		return
	}

	challenge, err := h.service.Challenge(c.Request.Context(), address)
	if err != nil {
		dto.RespondAppError(c, err)
		return
	}

	c.JSON(200, dto.NewSuccessResponse(gin.H{"challenge": challenge}))
}

func (h *RegisterHandler) Verify(c *gin.Context) {
	var req struct {
		Address   string `json:"address" binding:"required,eth_addr"`
		Signature string `json:"signature" binding:"required,len=132"`
		Nonce     string `json:"nonce" binding:"required"`
		Username  string `json:"username" binding:"required,min=3,max=50"`
	}

	if !dto.BindJSON(c, &req) {
		return
	}

	valid, err := h.service.Verify(c.Request.Context(), req.Address, req.Nonce, req.Signature, req.Username)
	if err != nil {
		dto.RespondAppError(c, err)
		return
	}

	setAuthCookie(c, "auth_token", valid, 24*3600)

	c.JSON(200, dto.NewSuccessResponse(gin.H{"valid": true}))
}

// Logout menghapus cookie auth_token dan menambahkan token ke Redis
// blacklist. Token yang sudah di-blacklist tidak bisa dipakai lagi
// meskipun belum expired (sampai TTL 24 jam habis).
func (h *RegisterHandler) Logout(c *gin.Context) {
	token, _ := c.Cookie("auth_token")
	blacklistToken(c, token, h.memory)
	setAuthCookie(c, "auth_token", "", -3600)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Logged out successfully"))
}
