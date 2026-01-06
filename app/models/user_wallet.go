package models

type UserWallet struct {
	UserAddress      string  `db:"user_address"`
	YTEBalance       float64 `db:"yte_balance"`
	LockedBalance    float64 `db:"locked_balance"`
	AvailableBalance float64 `db:"available_balance"`
	TotalReceived    float64 `db:"total_received"`
	TotalSent        float64 `db:"total_sent"`
	LastTransaction  string  `db:"last_transaction_at"`
}

type WalletHistory struct {
	UserAddress   string  `db:"user_address"`
	TxID          *int64  `db:"tx_id"`
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
