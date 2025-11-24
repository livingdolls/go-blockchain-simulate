package utils

const (
	MinimumFee      = 0.001 // Minimum transaction fee
	LowAmountFee    = 0.001 // for amounts < 10
	MediumAmountFee = 0.01  // for amounts between 10 and 100
	HighAmountRate  = 0.001 // 0.1% for amounts >= 100
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

// validate checks
func ValidateTransactionFee(amount, providedFee float64) bool {
	minimumRequiredFee := CalculateTransactionFee(amount)
	return providedFee >= minimumRequiredFee
}

// format fee to 8 decimal places
func FormatFee(fee float64) float64 {
	return float64(int64(fee*100000000)) / 100000000
}
