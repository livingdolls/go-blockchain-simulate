package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBlockReward(t *testing.T) {
	tests := []struct {
		name        string
		blockNumber int64
		want        float64
	}{
		{"block 0 (era 0)", 0, 50.0},
		{"block 1 (era 0)", 1, 50.0},
		{"block 99 (last of era 0)", 99, 50.0},
		{"block 100 (era 1)", 100, 25.0},
		{"block 101 (era 1)", 101, 25.0},
		{"block 199 (last of era 1)", 199, 25.0},
		{"block 200 (era 2)", 200, 12.5},
		{"block 300 (era 3)", 300, 6.25},
		{"block 400 (era 4)", 400, 3.125},
		{"block 1000 (era 10)", 1000, 50.0 / 1024},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateBlockReward(tt.blockNumber)
			assert.InDelta(t, tt.want, got, 1e-9,
				"reward block %d harus %v, got %v", tt.blockNumber, tt.want, got)
		})
	}
}

func TestCalculateBlockReward_FloorAtMinimum(t *testing.T) {
	// Setelah ~27 halvings, reward < MinimumReward -> floor
	// 2^27 = 134217728, halvings ke 27 dimulai di block 27*100=2700
	// reward di block 2700 = 50 / 2^27 = ~3.7e-7 < MinimumReward (1e-8)? No, 3.7e-7 > 1e-8
	// Mari coba block yang lebih jauh
	// 2^30 = 1.07e9, reward 50/1.07e9 = 4.66e-8 > 1e-8
	// 2^33 = 8.59e9, reward 50/8.59e9 = 5.82e-9 < 1e-8
	// blockNumber = 33*100 = 3300 -> halvings = 33, reward = 50/2^33
	// Actually: blockNumber/100 = 33, halvings = 33, reward = 50/2^33
	// 2^33 = 8589934592
	// 50 / 8589934592 = 5.82e-9 < 1e-8
	got := CalculateBlockReward(3300)
	assert.Equal(t, MinimumReward, got,
		"reward setelah ~33 halvings harus floor di MinimumReward (%v), got %v", MinimumReward, got)
}

func TestGetCurrentSupply(t *testing.T) {
	tests := []struct {
		name        string
		blockNumber int64
		want        float64
	}{
		// Era 0: blocks 1-100, reward 50 each = 5000
		{"block 0 -> 0", 0, 0},
		{"block 1 -> 50", 1, 50},
		{"block 100 -> 5000 (full era 0)", 100, 5000},
		// Era 1: blocks 101-200, reward 25 each = 2500; total 7500
		{"block 200 -> 7500 (full era 0+1)", 200, 7500},
		// Era 2: blocks 201-300, reward 12.5; 50 blocks * 12.5 = 625; total 8125
		{"block 250 -> 8125 (per docs example)", 250, 8125},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, GetCurrentSupply(tt.blockNumber), 1e-6)
		})
	}
}

func TestGetMaxSupply(t *testing.T) {
	// Geometris deret: 100 * 50 * (1 - (1/2)^20) / (1 - 1/2) = 10000 * (1 - 9.5e-7) ≈ 10000
	got := GetMaxSupply()
	assert.InDelta(t, 9999.99023, got, 0.01,
		"max supply konvergen ke 10000 (geometric series), got %v", got)
}

func TestGetNextHalvingBlock(t *testing.T) {
	tests := []struct {
		in   int64
		want int64
	}{
		{0, 100},
		{1, 100},
		{99, 100},
		{100, 200},
		{250, 300},
		{300, 400},
		{999, 1000},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, GetNextHalvingBlock(tt.in))
		})
	}
}

func TestGetBlocksUntilHalving(t *testing.T) {
	tests := []struct {
		in   int64
		want int64
	}{
		{0, 100},
		{99, 1},
		{100, 100},
		{250, 50},
		{300, 100},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, GetBlocksUntilHalving(tt.in))
		})
	}
}
