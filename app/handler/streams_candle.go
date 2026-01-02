package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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
	log.Printf("SSE connected for interval: %s\n", interval)

	// write initial candle
	go func() {
		latestCandle, err := h.candleService.GetLatestCandleByInterval(interval)
		if err == nil {
			data, err := json.Marshal(latestCandle)

			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				flusher.Flush()
				log.Printf("Initial candle sent for interval: %s\n", interval)
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
				log.Printf("Marshal candle error: %v\n", err)
				return err
			}

			// recover panic
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered in candle stream SSE: %v", r)
					cancel()
				}
			}()

			// safe write

			if _, err := fmt.Fprintf(w, "data: %s\n\n", string(data)); err != nil {
				log.Printf("Write to SSE error: %v\n", err)
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
		log.Printf("SSE disconnected for interval: %s\n", interval)
		cancel()
	case err := <-errChan:
		log.Printf("SSE error for interval %s: %v\n", interval, err)
		cancel()
	}

	<-doneChan
	log.Printf("SSE routine ended for interval: %s\n", interval)
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
			log.Println("SSE ping disconnected")
			return
		}
	}
}
