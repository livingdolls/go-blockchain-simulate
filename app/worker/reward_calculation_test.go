package worker

import (
	"testing"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/stretchr/testify/assert"
)

// newTestConsumer membuat RewardCalculationConsumer dengan dependencies
// minimal untuk test calculationBonusReward (method pure, tidak pakai state).
func newTestConsumer() *RewardCalculationConsumer {
	return &RewardCalculationConsumer{
		processed:     make(map[int64]ProcessedBlock),
		pendingBlocks: make(map[int64]dto.RewardCalculationEvent),
	}
}

func TestCalculationBonusReward_NoTx(t *testing.T) {
	// < 10 tx: tx bonus 0. No fees: miner bonus 0. Total: 0
	c := newTestConsumer()
	bonus := c.calculationBonusReward(dto.RewardCalculationEvent{
		BlockReward:         50.0,
		TransactionCount:    0,
		MinerAddress:        "MINER",
		TotalTransactionFee: 0,
	})
	assert.Equal(t, 0.0, bonus)
}

func TestCalculationBonusReward_BelowTxThreshold(t *testing.T) {
	// 9 tx: tx bonus 0 (di bawah threshold 10)
	c := newTestConsumer()
	bonus := c.calculationBonusReward(dto.RewardCalculationEvent{
		BlockReward:         50.0,
		TransactionCount:    9,
		MinerAddress:        "MINER",
		TotalTransactionFee: 1.0,
	})
	// tx bonus 0, miner bonus 0.5%, total 0.5% * 50 = 0.25
	assert.InDelta(t, 0.25, bonus, 1e-9)
}

func TestCalculationBonusReward_AtTxThreshold(t *testing.T) {
	// 10 tx: tx bonus 10/10=1%. + 0.5% miner. Total 1.5%
	c := newTestConsumer()
	bonus := c.calculationBonusReward(dto.RewardCalculationEvent{
		BlockReward:         100.0,
		TransactionCount:    10,
		MinerAddress:        "MINER",
		TotalTransactionFee: 1.0,
	})
	// 1.5% * 100 = 1.5
	assert.InDelta(t, 1.5, bonus, 1e-9)
}

func TestCalculationBonusReward_TxBonusCappedAtFive(t *testing.T) {
	// 50 tx: tx bonus 50/10=5% (capped, max 5%). + 0.5% miner. Total 5.5%
	c := newTestConsumer()
	bonus := c.calculationBonusReward(dto.RewardCalculationEvent{
		BlockReward:         100.0,
		TransactionCount:    50,
		MinerAddress:        "MINER",
		TotalTransactionFee: 1.0,
	})
	// 5.5% * 100 = 5.5
	assert.InDelta(t, 5.5, bonus, 1e-9)
}

func TestCalculationBonusReward_TxBonusExceedsFive(t *testing.T) {
	// 100 tx: tx bonus 100/10=10% di-cap ke 5%. + 0.5% miner = 5.5%
	c := newTestConsumer()
	bonus := c.calculationBonusReward(dto.RewardCalculationEvent{
		BlockReward:         100.0,
		TransactionCount:    100,
		MinerAddress:        "MINER",
		TotalTransactionFee: 1.0,
	})
	assert.InDelta(t, 5.5, bonus, 1e-9)
}

func TestCalculationBonusReward_TotalCapAtTen(t *testing.T) {
	// 1000 tx: tx bonus capped 5% + miner 0.5% = 5.5% (di bawah 10%, ok)
	// 2000 tx: hypothetical tidak bisa, karena tx bonus di-cap 5%.
	// Verifikasi total cap di 10% via extreme case: bayangkan
	// MinerAddress empty + tx 0 -> 0%. Plus edge case:
	// Bonus tidak pernah bisa cap 10% dengan rule saat ini.
	// Test ini untuk mengunci invariant: max bonus = 5% (tx) + 0.5% (miner) = 5.5%
	c := newTestConsumer()
	bonus := c.calculationBonusReward(dto.RewardCalculationEvent{
		BlockReward:         1.0,
		TransactionCount:    1000,
		MinerAddress:        "MINER",
		TotalTransactionFee: 100.0,
	})
	// Maks = 5.5% dari BlockReward
	assert.LessOrEqual(t, bonus, 0.055,
		"bonus tidak boleh pernah > 5.5%% BlockReward (aturan saat ini)")
}

func TestCalculationBonusReward_NoMinerBonusWhenNoFees(t *testing.T) {
	// Tx 50 tapi fee=0 dan miner address kosong -> no miner bonus
	c := newTestConsumer()
	bonus := c.calculationBonusReward(dto.RewardCalculationEvent{
		BlockReward:         100.0,
		TransactionCount:    50,
		MinerAddress:        "",
		TotalTransactionFee: 0,
	})
	// 5% * 100 = 5
	assert.InDelta(t, 5.0, bonus, 1e-9)
}

func TestGetMetrics_EmptyConsumer(t *testing.T) {
	c := newTestConsumer()
	m := c.GetMetrics()
	assert.Equal(t, int64(0), m["processed_count"])
	assert.Equal(t, int64(0), m["failed_count"])
	assert.Equal(t, int64(0), m["retried_count"])
	assert.Equal(t, 0.0, m["total_rewards_issued"])
	assert.Equal(t, 0.0, m["avg_reward_per_block"], "avg reward = 0 ketika tidak ada processed")
}

func TestGetMetrics_AfterSuccess(t *testing.T) {
	c := newTestConsumer()
	c.recordSuccess(50.0, 100.0)
	c.recordSuccess(25.0, 50.0)
	c.recordFailure()

	m := c.GetMetrics()
	assert.Equal(t, int64(2), m["processed_count"])
	assert.Equal(t, int64(1), m["failed_count"])
	assert.Equal(t, 75.0, m["total_rewards_issued"])
	assert.Equal(t, 150.0, m["total_usd_value_issued"])
	assert.Equal(t, 37.5, m["avg_reward_per_block"]) // 75/2
}
