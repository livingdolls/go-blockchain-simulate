package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
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

type RegisterHandler struct {
	service services.RegisterService
}

func NewRegisterHandler(service services.RegisterService) *RegisterHandler {
	return &RegisterHandler{service: service}
}

func (h *RegisterHandler) Register(c *gin.Context) {
	var req models.UserRegister
	if !dto.BindJSON(c, &req) {
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Ethereum address format"})
		return
	}

	challenge, err := h.service.Challenge(c.Request.Context(), address)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"challenge": challenge})
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
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	setAuthCookie(c, "auth_token", valid, 24*3600)

	c.JSON(200, gin.H{"valid": true})
}
