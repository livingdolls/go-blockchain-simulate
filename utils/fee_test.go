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
