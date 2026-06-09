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
	Address     string  `json:"address" binding:"required,eth_addr"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ReferenceID string  `json:"reference_id,omitempty" binding:"omitempty,max=100"`
	Description string  `json:"description,omitempty" binding:"omitempty,max=500"`
}
