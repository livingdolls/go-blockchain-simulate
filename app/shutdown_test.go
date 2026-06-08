package app

import (
	"testing"

	"github.com/livingdolls/go-blockchain-simulate/app/worker"
)

// Compile-time assertion: setiap worker/consumer yang dipakai di
// shutdown.go HARUS implement interface stoppable. Jika ada worker
// baru yang ditambahkan tapi lupa menambahkan method Stop(), baris
// `var _ stoppable = ...` akan GAGAL COMPILE.
//
// Test ini tidak punya runtime assertion; fungsinya murni
// mengunci kontrak saat kompilasi.
func TestAllShutdownWorkersImplementStoppable(t *testing.T) {
	var _ stoppable = (*worker.GenerateBlockWorker)(nil)
	var _ stoppable = (*worker.GenerateCandleWorker)(nil)
	var _ stoppable = (*worker.TransactionConsumer)(nil)
	var _ stoppable = (*worker.MarketPricingConsumer)(nil)
	var _ stoppable = (*worker.MarketVolumeConsumer)(nil)
	var _ stoppable = (*worker.LedgerAuditConsumer)(nil)
	var _ stoppable = (*worker.LedgerReconcileConsumer)(nil)
	var _ stoppable = (*worker.RewardCalculationConsumer)(nil)
	var _ stoppable = (*worker.RewardDistributionConsumer)(nil)
	var _ stoppable = (*worker.NotificationWSConsumer)(nil)
	// Test pass jika baris di atas compile (yaitu setiap struct
	// mengimplementasikan Stop()).
}
