package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateTransactionFee(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   float64
	}{
		{"zero amount -> minimum", 0, MinimumFee},
		{"negative amount -> minimum", -100, MinimumFee},
		{"amount 0.5 (small) -> minimum", 0.5, MinimumFee},
		{"amount 9.99 (just below 10) -> minimum", 9.99, MinimumFee},
		{"amount 10 (just at 10) -> medium", 10, MediumAmountFee},
		{"amount 50 (medium) -> medium", 50, MediumAmountFee},
		{"amount 99.99 (just below 100) -> medium", 99.99, MediumAmountFee},
		{"amount 100 (just at 100) -> 0.1%", 100, 0.1},
		{"amount 1000 -> 0.1%", 1000, 1.0},
		{"amount 10000 -> 0.1%", 10000, 10.0},
		{"amount 50000 -> 0.1%", 50000, 50.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CalculateTransactionFee(tt.amount),
				"CalculateTransactionFee(%v) = %v, want %v", tt.amount, CalculateTransactionFee(tt.amount), tt.want)
		})
	}
}

func TestValidateTransactionFee(t *testing.T) {
	tests := []struct {
		name string
		amt  float64
		fee  float64
		want bool
	}{
		{"exact minimum for small amount", 5, MinimumFee, true},
		{"below minimum rejected", 5, MinimumFee - 0.0001, false},
		{"exact medium for medium amount", 50, MediumAmountFee, true},
		{"below medium rejected", 50, 0.005, false},
		{"exact 0.1% for large amount", 1000, 1.0, true},
		{"slightly below 0.1% rejected", 1000, 0.99, false},
		{"above minimum accepted", 50, MediumAmountFee + 0.5, true},
		{"zero fee rejected for any amount", 100, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidateTransactionFee(tt.amt, tt.fee))
		})
	}
}

func TestFormatFee(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"zero", 0, 0},
		{"already 8 decimals", 0.12345678, 0.12345678},
		{"truncate excess decimals", 0.123456789, 0.12345678},
		{"rounding down at 9th decimal", 0.000000009, 0},
		{"large value preserved to 8 decimals", 12345.67890123, 12345.67890123},
		{"integer fee", 5.0, 5.0},
		{"minimum fee unchanged", MinimumFee, MinimumFee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatFee(tt.in))
		})
	}
}

func TestCalculateCongestionMultiplier(t *testing.T) {
	tests := []struct {
		name   string
		count  int
		expect float64
	}{
		{"0 pending", 0, MultiplierNone},
		{"10 pending (low)", 10, MultiplierNone},
		{"11 pending (medium)", 11, MultiplierLow},
		{"50 pending (medium)", 50, MultiplierLow},
		{"51 pending (high)", 51, MultiplierMedium},
		{"100 pending (high)", 100, MultiplierMedium},
		{"101 pending (very high)", 101, MultiplierHigh},
		{"200 pending (very high)", 200, MultiplierHigh},
		{"201 pending (extreme)", 201, MultiplierVeryHigh},
		{"1000 pending (extreme)", 1000, MultiplierVeryHigh},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, CalculateCongestionMultiplier(tt.count))
		})
	}
}

func TestEstimateFee(t *testing.T) {
	// Low congestion, low priority
	fee := EstimateFee(100, 5, PriorityLow)
	assert.Equal(t, 0.1, fee) // base=0.1, congestion=1.0, priority=1.0

	// Low congestion, high priority
	fee = EstimateFee(100, 5, PriorityHigh)
	assert.Equal(t, 0.2, fee) // base=0.1, congestion=1.0, priority=2.0

	// High congestion, low priority
	fee = EstimateFee(100, 150, PriorityLow)
	assert.Equal(t, 0.2, fee) // base=0.1, congestion=2.0, priority=1.0

	// High congestion, high priority
	fee = EstimateFee(100, 150, PriorityHigh)
	assert.Equal(t, 0.4, fee) // base=0.1, congestion=2.0, priority=2.0

	// Small amount, no congestion
	fee = EstimateFee(5, 0, PriorityLow)
	assert.Equal(t, MinimumFee, fee) // base=0.001, congestion=1.0, priority=1.0

	// Priority < 1.0 should be clamped to 1.0
	fee = EstimateFee(100, 0, 0.5)
	assert.Equal(t, 0.1, fee) // priority clamped to 1.0
}

func TestEstimatePriorityMultiplier(t *testing.T) {
	assert.Equal(t, PriorityLow, EstimatePriorityMultiplier("low"))
	assert.Equal(t, PriorityMedium, EstimatePriorityMultiplier("medium"))
	assert.Equal(t, PriorityHigh, EstimatePriorityMultiplier("high"))
	assert.Equal(t, PriorityLow, EstimatePriorityMultiplier(""))
	assert.Equal(t, PriorityLow, EstimatePriorityMultiplier("unknown"))
}

func TestCongestionLevel(t *testing.T) {
	assert.Equal(t, "low", CongestionLevel(0))
	assert.Equal(t, "low", CongestionLevel(10))
	assert.Equal(t, "medium", CongestionLevel(11))
	assert.Equal(t, "medium", CongestionLevel(50))
	assert.Equal(t, "high", CongestionLevel(51))
	assert.Equal(t, "high", CongestionLevel(100))
	assert.Equal(t, "very_high", CongestionLevel(101))
	assert.Equal(t, "very_high", CongestionLevel(500))
}

func TestCongestionPercentage(t *testing.T) {
	assert.Equal(t, 0.0, CongestionPercentage(0))
	assert.Equal(t, 50.0, CongestionPercentage(100))
	assert.Equal(t, 100.0, CongestionPercentage(200))
	// Past 200 should be capped at 100
	assert.Equal(t, 100.0, CongestionPercentage(500))
}
