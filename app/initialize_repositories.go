package app

import (
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// InitializeRepositories menginisialisasi semua repository yang dipakai
// oleh service layer. Urutan tidak penting; setiap repository independen.
func (a *AppConfig) InitializeRepositories() {
	a.UserRepo = repository.NewUserRepository(a.DB)
	a.WalletRepo = repository.NewUserWalletRepository(a.DB)
	a.BalanceRepo = repository.NewUserBalanceRepository(a.DB)
	a.TxRepo = repository.NewTransactionRepository(a.DB)
	a.LedgerRepo = repository.NewLedgerRepository(a.DB)
	a.MarketRepo = repository.NewMarketRepository(a.DB)
	a.BlockRepo = repository.NewBlockRepository(a.DB)
	a.CandleRepo = repository.NewCandleRepository(a.DB)
	a.DiscrepancyRepo = repository.NewDiscrepancyRepository(a.DB)
	a.AdminRepo = repository.NewAdminRepository(a.DB)
	logger.LogInfo("All repositories initialized successfully")
}
