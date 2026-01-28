package dto

type LedgerEntryEvent struct {
	EntryID      int64   `json:"entry_id"`
	TxID         *int64  `json:"tx_id"`
	BlockID      int64   `json:"block_id"`
	BlockNumber  int     `json:"block_number"`
	Address      string  `json:"address"`
	Amount       float64 `json:"amount"`
	BalanceAfter float64 `json:"balance_after"`
	EntryType    string  `json:"entry_type"`
	Timestamp    int64   `json:"timestamp"`
	CreatedAt    string  `json:"created_at"`
}

type LedgerBatchEvent struct {
	BlockID      int64              `json:"block_id"`
	BlockNumber  int                `json:"block_number"`
	TotalEntries int                `json:"total_entries"`
	Entries      []LedgerEntryEvent `json:"entries"`
	Timestamp    int64              `json:"timestamp"`
	MinerAddress string             `json:"miner_address"`
}

type BalanceReconciliation struct {
	Address         string  `json:"address"`
	ExpectedBalance float64 `json:"expected_balance"`
	ActualBalance   float64 `json:"actual_balance"`
	Difference      float64 `json:"difference"`
	LastEntryID     int64   `json:"last_entry_id"`
	BlockNumber     int     `json:"block_number"`
	Timestamp       int64   `json:"timestamp"`
}

type AuditTrailEntry struct {
	EntryID     int64   `json:"entry_id"`
	Action      string  `json:"action"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      float64 `json:"amount"`
	Fee         float64 `json:"fee"`
	BlockNumber int     `json:"block_number"`
	Timestamp   int64   `json:"timestamp"`
	Reconciled  bool    `json:"reconciled"`
}
