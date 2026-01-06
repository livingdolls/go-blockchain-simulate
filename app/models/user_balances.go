package models

type UserBalance struct {
	UserAddress     string  `db:"user_address"`
	USDBalance      float64 `db:"usd_balance"`
	LockedBalance   float64 `db:"locked_balance"`
	TotalDeposited  float64 `db:"total_deposited"`
	TotalWithdrawn  float64 `db:"total_withdrawn"`
	TotalTraded     float64 `db:"total_traded"`
	LastTransaction string  `db:"last_transaction_at"`
}

type BalanceHistory struct {
	UserAddress   string  `db:"user_address"`
	OrderID       *int64  `db:"order_id"`
	ChangeType    string  `db:"change_type"`
	Amount        float64 `db:"amount"`
	BalanceBefore float64 `db:"balance_before"`
	BalanceAfter  float64 `db:"balance_after"`
	LockedBefore  float64 `db:"locked_before"`
	LockedAfter   float64 `db:"locked_after"`
	ReferenceID   *string `db:"reference_id"`
	Description   *string `db:"description"`
}
