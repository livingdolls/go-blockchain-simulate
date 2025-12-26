package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type MarketHandler struct {
	service services.MarketEngineService
}

func NewMarketHandler(service services.MarketEngineService) *MarketHandler {
	return &MarketHandler{
		service: service,
	}
}

func (h *MarketHandler) GetMarketEngineState(c *gin.Context) {
	state, err := h.service.GetState()
	if err != nil {
		c.JSON(500, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(200, dto.NewSuccessResponse(state))
}
