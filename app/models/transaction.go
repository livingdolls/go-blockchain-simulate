package models

import "strings"

type Transaction struct {
	ID          int64   `db:"id" json:"id"`
	FromAddress string  `db:"from_address" json:"from_address"`
	ToAddress   string  `db:"to_address" json:"to_address"`
	Amount      float64 `db:"amount" json:"amount"`
	Fee         float64 `db:"fee" json:"fee"`
	Type        string  `db:"type" json:"type"` // "TRANSFER", "BUY", "SELL"
	Signature   string  `db:"signature" json:"signature"`
	Status      string  `db:"status" json:"status"`
}

type TransactionWithType struct {
	Transaction
	TypeTx string `db:"type" json:"type"` // "sent" or "received"
}

type TransactionFilter struct {
	Address string `json:"address"`
	Type    string `json:"type"`   // "all", "sent", "received"
	Status  string `json:"status"` // "all", "pending", "confirmed"
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	SortBy  string `json:"sort_by"` // "id", "amount", "created_at"
	Order   string `json:"order"`   // "asc", "desc"
}

type TransactionWithTypeResponse struct {
	Transactions []TransactionWithType `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	Limit        int                   `json:"limit"`
	TotalPages   int                   `json:"total_pages"`
}

// Validate and set defaults
func (f *TransactionFilter) Validate() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 10
	}

	//normalize type to lowercase
	f.Type = strings.ToLower(strings.TrimSpace(f.Type))
	if f.Type == "" {
		f.Type = "all"
	}
	// validate type, only allow "all", "send", "received"
	validTypes := map[string]bool{"all": true, "send": true, "received": true}
	if !validTypes[f.Type] {
		f.Type = "all"
	}

	// validate status, only allow "all", "pending", "confirmed"
	f.Status = strings.ToUpper(strings.TrimSpace(f.Status))
	if f.Status == "" {
		f.Status = "ALL"
	}
	validStatuses := map[string]bool{"ALL": true, "PENDING": true, "CONFIRMED": true}
	if !validStatuses[f.Status] {
		f.Status = "ALL"
	}

	// Validate sort
	validSorts := map[string]bool{"id": true, "amount": true}
	if !validSorts[f.SortBy] {
		f.SortBy = "id"
	}

	// Validate order
	f.Order = strings.ToUpper(f.Order)
	if f.Order != "ASC" && f.Order != "DESC" {
		f.Order = "DESC"
	}
}
