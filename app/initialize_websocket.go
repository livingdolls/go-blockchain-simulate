package app

import (
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// InitializeWebSocket menginisialisasi WebSocket hub dan publisher-nya.
// Hub.Run() berjalan di goroutine terpisah; hub.Close() dipanggil saat
// shutdown dari AppConfig.Shutdown().
func (a *AppConfig) InitializeWebSocket() {
	a.Hub = websocket.NewHub()
	go a.Hub.Run()
	a.PublisherWS = publisher.NewPublisherWS(a.Hub)
	logger.LogInfo("WebSocket hub initialized successfully")
}
