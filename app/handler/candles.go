package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type CandleHandler struct {
	service services.CandleService
}

func NewCandleHandler(service services.CandleService) *CandleHandler {
	return &CandleHandler{
		service: service,
	}
}

func (h *CandleHandler) GetCandle(c *gin.Context) {
	intervalType := c.Query("interval") // '1m', '5m', '15m', '30m', '1h', '4h', '1d'
	limit := c.DefaultQuery("limit", "100")

	// validation intervalType
	if !validateIntervalType(intervalType) {
		c.JSON(400, dto.NewErrorResponse[string]("invalid interval type"))
		return
	}

	// convert limit to int
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(400, dto.NewErrorResponse[string]("invalid limit"))
		return
	}

	candles, err := h.service.GetCandles(intervalType, limitInt)
	if err != nil {
		c.JSON(500, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(200, dto.NewSuccessResponse(candles))
}

func (h *CandleHandler) GetCandleFrom(c *gin.Context) {
	intervalType := c.Query("interval") // '1m', '5m', '15m', '30m', '1h', '4h', '1d'
	startTimeStr := c.Query("start_time")
	limit := c.DefaultQuery("limit", "100")

	// validation intervalType
	if !validateIntervalType(intervalType) {
		c.JSON(400, dto.NewErrorResponse[string]("invalid interval type"))
		return
	}

	startTimeInt, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil || startTimeInt <= 0 {
		c.JSON(400, dto.NewErrorResponse[string]("invalid start_time"))
		return
	}

	// convert limit to int
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(400, dto.NewErrorResponse[string]("invalid limit"))
		return
	}

	candles, err := h.service.GetCandlesFrom(intervalType, startTimeInt, limitInt)
	if err != nil {
		c.JSON(500, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(200, dto.NewSuccessResponse(candles))
}

func validateIntervalType(intervalType string) bool {
	validIntervals := map[string]bool{
		"1m":  true,
		"5m":  true,
		"15m": true,
		"30m": true,
		"1h":  true,
		"4h":  true,
		"1d":  true,
	}

	return validIntervals[intervalType]
}
