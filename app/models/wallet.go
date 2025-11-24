package models

type WalletResponse struct {
	Ballance     float64    `json:"balance"`
	Address      string     `json:"address"`
	Transactions []WalletTx `json:"transactions"`
}

type WalletTx struct {
	ID     int64   `json:"id"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Fee    float64 `json:"fee"`
	Status string  `json:"status"`
	Type   string  `json:"type"` // "send" atau "receive"
}
