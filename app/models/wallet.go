package models

type WalletResponse struct {
	Address      string                      `json:"address"`
	Transactions TransactionWithTypeResponse `json:"transactions"`
}
