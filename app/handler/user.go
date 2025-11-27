package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

type UserHandler struct {
	svc services.ProfileService
	jwt security.JWTService
}

func NewUserHandler(svc services.ProfileService, jwt security.JWTService) *UserHandler {
	return &UserHandler{
		svc: svc,
		jwt: jwt,
	}
}

func (h *UserHandler) Me(c *gin.Context) {
	claims, ok := GetUserClaims(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	address := claims.Address

	user, err := h.svc.Me(address)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"user": user})
}
