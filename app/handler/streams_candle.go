package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type CandleStreamHandler struct {
	streamService services.CandleStreamService
	candleService services.CandleService
}

func NewCandleStreamHandler(streamService services.CandleStreamService, candleService services.CandleService) *CandleStreamHandler {
	return &CandleStreamHandler{
		streamService: streamService,
		candleService: candleService,
	}
}

func (h *CandleStreamHandler) StreamCandles(c *gin.Context) {
	interval := c.Query("interval")

	if !dto.IsValidInterval(interval) {
		c.JSON(400, dto.NewErrorResponse[string]("invalid interval"))
		return
	}

	// set SSE headers
	utils.SetupSSEHeaders(c)

	w := c.Writer
	flusher, ok := w.(http.Flusher)

	if !ok {
		c.JSON(500, dto.NewErrorResponse[string]("streaming unsupported"))
		return
	}

	// create context
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// listen untuk client
	logger.LogInfo("SSE connected for interval: " + interval)

	// write initial candle
	go func() {
		latestCandle, err := h.candleService.GetLatestCandleByInterval(interval)
		if err == nil {
			data, err := json.Marshal(latestCandle)

			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				flusher.Flush()
				logger.LogInfo("Initial candle sent for interval: " + interval)
			}
		}
	}()

	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// subscribe to candle stream
	go func() {
		defer close(doneChan)
		err := h.streamService.SubscribeCandle(ctx, interval, func(candle models.Candle) error {
			// check context

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			data, err := json.Marshal(candle)

			if err != nil {
				logger.LogError("Marshal candle error", err)
				return err
			}

			// recover panic
			defer func() {
				if r := recover(); r != nil {
					logger.LogInfo(fmt.Sprintf("Recovered in candle stream SSE: %v", r))
				}
			}()
			// safe write

			if _, err := fmt.Fprintf(w, "data: %s\n\n", string(data)); err != nil {
				logger.LogError("Write to SSE error", err)
				return err
			}

			flusher.Flush()
			return nil
		})

		if err != nil && err != context.Canceled {
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	// handle disconnection
	select {
	case <-c.Request.Context().Done():
		logger.LogInfo("SSE disconnected for interval: " + interval)
		cancel()
	case err := <-errChan:
		logger.LogError("SSE error for interval "+interval, err)
		cancel()
	}

	<-doneChan
	logger.LogInfo("SSE routine ended for interval: " + interval)
}

func (h *CandleStreamHandler) Ping(c *gin.Context) {
	utils.SetupSSEHeaders(c)

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(500, dto.NewErrorResponse[string]("streaming unsupported"))
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(w, "data: %s\n\n", "ping")
			flusher.Flush()
		case <-c.Request.Context().Done():
			logger.LogInfo("SSE ping disconnected")
			return
		}
	}
}
