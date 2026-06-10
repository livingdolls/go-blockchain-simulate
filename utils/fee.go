package utils

import "math"

const (
	MinimumFee      = 0.001 // Minimum transaction fee
	LowAmountFee    = 0.001 // for amounts < 10
	MediumAmountFee = 0.01  // for amounts between 10 and 100
	HighAmountRate  = 0.001 // 0.1% for amounts >= 100

	// Congestion thresholds (pending transaction count)
	CongestionLow      = 10
	CongestionMedium   = 50
	CongestionHigh     = 100
	CongestionVeryHigh = 200

	// Congestion multipliers
	MultiplierNone     = 1.0
	MultiplierLow      = 1.2
	MultiplierMedium   = 1.5
	MultiplierHigh     = 2.0
	MultiplierVeryHigh = 3.0

	// Priority levels
	PriorityLow    = 1.0 // standard speed
	PriorityMedium = 1.5 // faster
	PriorityHigh   = 2.0 // fastest
)

func CalculateTransactionFee(amount float64) float64 {
	if amount <= 0 {
		return MinimumFee
	}

	// small transactions : fixed minimum fee
	if amount < 10 {
		return MinimumFee
	}

	// medium transactions : fixed medium fee
	if amount < 100 {
		return MediumAmountFee
	}

	// large transactions : percentage-based fee
	fee := amount * HighAmountRate
	if fee < MinimumFee {
		fee = MinimumFee
	}

	return fee
}

// CalculateCongestionMultiplier menghitung multiplier fee berdasarkan
// jumlah transaksi pending di network. Semakin banyak pending tx,
// semakin tinggi fee yang dibutuhkan untuk diproses duluan.
//
// Model: tiered multiplier berdasarkan threshold:
//   - 0-10 pending   -> 1.0x (normal)
//   - 11-50 pending  -> 1.2x
//   - 51-100 pending -> 1.5x
//   - 101-200 pending -> 2.0x
//   - 200+ pending   -> 3.0x
func CalculateCongestionMultiplier(pendingCount int) float64 {
	switch {
	case pendingCount <= CongestionLow:
		return MultiplierNone
	case pendingCount <= CongestionMedium:
		return MultiplierLow
	case pendingCount <= CongestionHigh:
		return MultiplierMedium
	case pendingCount <= CongestionVeryHigh:
		return MultiplierHigh
	default:
		return MultiplierVeryHigh
	}
}

// EstimateFee menghitung fee dengan mempertimbangkan congestion dan
// priority. Formula:
//
//	fee = baseFee * congestionMultiplier * priorityMultiplier
//
// Hasil di-clamp ke MinimumFee agar tidak pernah 0.
func EstimateFee(amount float64, pendingCount int, priorityMultiplier float64) float64 {
	baseFee := CalculateTransactionFee(amount)
	congestion := CalculateCongestionMultiplier(pendingCount)

	if priorityMultiplier < 1.0 {
		priorityMultiplier = PriorityLow
	}

	fee := baseFee * congestion * priorityMultiplier
	if fee < MinimumFee {
		fee = MinimumFee
	}

	return FormatFee(fee)
}

// EstimatePriorityMultiplier mengkonversi level string ke multiplier.
// Level: "low" (1.0x), "medium" (1.5x), "high" (2.0x).
func EstimatePriorityMultiplier(level string) float64 {
	switch level {
	case "high":
		return PriorityHigh
	case "medium":
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// CongestionLevel mengembalikan label congestion berdasarkan pending count.
func CongestionLevel(pendingCount int) string {
	switch {
	case pendingCount <= CongestionLow:
		return "low"
	case pendingCount <= CongestionMedium:
		return "medium"
	case pendingCount <= CongestionHigh:
		return "high"
	default:
		return "very_high"
	}
}

// CongestionPercentage mengembalikan persentase congestion (0-100)
// untuk ditampilkan di UI. Berdasarkan skala max 200 pending.
func CongestionPercentage(pendingCount int) float64 {
	pct := float64(pendingCount) / float64(CongestionVeryHigh) * 100
	return math.Min(pct, 100)
}

// ValidateTransactionFee validates that the provided fee meets the minimum
// required fee based on amount and congestion.
func ValidateTransactionFee(amount, providedFee float64) bool {
	minimumRequiredFee := CalculateTransactionFee(amount)
	return providedFee >= minimumRequiredFee
}

// FormatFee format fee to 8 decimal places
func FormatFee(fee float64) float64 {
	return float64(int64(fee*100000000)) / 100000000
}
