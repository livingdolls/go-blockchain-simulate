package dto

type TopUpResultDTO struct {
	Address       string  `json:"address"`
	Amount        float64 `json:"amount"`
	BalanceBefore float64 `json:"balance_before"`
	BalanceAfter  float64 `json:"balance_after"`
	ReferenceID   *string `json:"reference_id,omitempty"`
	Description   *string `json:"description,omitempty"`
}

type TopUpRequestDTO struct {
	Address     string  `json:"address"`
	Amount      float64 `json:"amount"`
	ReferenceID string  `json:"reference_id,omitempty"`
	Description string  `json:"description,omitempty"`
}
