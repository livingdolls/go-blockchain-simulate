package models

type WalletResponse struct {
	Ballance     float64                     `json:"balance"`
	Address      string                      `json:"address"`
	Transactions TransactionWithTypeResponse `json:"transactions"`
}
