package app

import (
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/config"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"go.uber.org/zap"
)

// NewAppDependencies membangun seluruh dependensi infrastruktur dari konfigurasi.
// Entry point untuk inisialisasi aplikasi. Dipanggil dari main sebelum
// inisialisasi handler/service/worker.
func NewAppDependencies(cfg *config.Config) (*AppDependencies, error) {
	deps := &AppDependencies{Config: cfg}

	if err := deps.initializeInfrastructure(); err != nil {
		return nil, err
	}

	return deps, nil
}

// InitializeInfrastructure tetap dipertahankan untuk backward compatibility.
// Memindahkan nilai dari AppDependencies ke AppConfig.
// Disarankan migrasi ke SetDeps di caller.
func (a *AppConfig) InitializeInfrastructure() error {
	if a.deps == nil {
		return entity.ErrDependenciesNotInitialized
	}
	a.copyDepsToConfig()
	return nil
}

// SetDeps menempelkan AppDependencies ke AppConfig dan langsung menyalin
// field-field infrastruktur. Cara yang disarankan di main.
func (a *AppConfig) SetDeps(deps *AppDependencies) {
	a.deps = deps
	a.copyDepsToConfig()
}

// copyDepsToConfig menyalin field infrastruktur dari deps ke AppConfig.
// Dipakai oleh InitializeInfrastructure dan SetDeps.
func (a *AppConfig) copyDepsToConfig() {
	if a.deps == nil {
		return
	}
	a.DBConn = a.deps.DBConn
	a.DB = a.deps.DBConn.GetDB()
	a.RedisServices = a.deps.RedisServices
	a.RMQClient = a.deps.RMQClient
	a.JWT = a.deps.JWT
	a.JWTAdmin = a.deps.JWTAdmin
	logger.LogInfo("AppConfig dependencies siap dipakai")
}

// SetupRabbitMQTopology mendeklarasikan semua queue, exchange, dan binding
// yang dibutuhkan aplikasi. Aman dipanggil setelah RabbitMQ client siap.
// Error di-deklarasi di-aggregate dan dikembalikan agar caller (main) bisa
// exit non-nol jika topology gagal di-setup (publisher akan silent drop
// kalau exchange/queue tidak ada di broker).
func (a *AppConfig) SetupRabbitMQTopology() error {
	if a.RMQClient == nil {
		return fmt.Errorf("RMQ client belum diinisialisasi")
	}

	queues := getQueueDefinitions()
	exchanges := getExchangeDefinitions()
	binds := getBindingDefinitions()

	var errs []string

	for _, q := range queues {
		if err := a.RMQClient.DeclareQueue(q); err != nil {
			logger.LogError("Failed to declare queue", err, zap.String("queue", q.Name))
			errs = append(errs, fmt.Sprintf("queue %s: %v", q.Name, err))
		}
	}

	for _, e := range exchanges {
		if err := a.RMQClient.DeclareExchange(e); err != nil {
			logger.LogError("Failed to declare exchange", err, zap.String("exchange", e.Name))
			errs = append(errs, fmt.Sprintf("exchange %s: %v", e.Name, err))
		}
	}

	for _, b := range binds {
		if err := a.RMQClient.Bind(b); err != nil {
			logger.LogError("Failed to bind", err, zap.String("queue", b.Queue), zap.String("exchange", b.Exchange))
			errs = append(errs, fmt.Sprintf("bind %s->%s: %v", b.Exchange, b.Queue, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("topology setup gagal (%d error): %v", len(errs), errs)
	}

	logger.LogInfo("RabbitMQ topology initialized successfully")
	return nil
}
