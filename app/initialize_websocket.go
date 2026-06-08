package app

import (
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// InitializeWebSocket menginisialisasi WebSocket hub dan publisher-nya.
// Hub.Run() berjalan di goroutine terpisah; kita tunggu Ready() signal
// agar PublisherWS.Publish() pertama dijamin tidak block (race window
// di mana Run() belum consume register/unregister channel).
// hub.Close() dipanggil saat shutdown dari AppConfig.Shutdown().
func (a *AppConfig) InitializeWebSocket() {
	a.Hub = websocket.NewHub()
	go a.Hub.Run()

	// Tunggu Run() loop aktif dengan timeout 5 detik. Idealnya selesai
	// dalam microseconds; timeout hanya defensive.
	select {
	case <-a.Hub.Ready():
		// Hub siap consume channel
	case <-time.After(5 * time.Second):
		logger.LogInfo("WebSocket hub ready timeout, lanjut inisialisasi (mungkin ada race window kecil)")
	}

	a.PublisherWS = publisher.NewPublisherWS(a.Hub)
	logger.LogInfo("WebSocket hub initialized successfully")
}
