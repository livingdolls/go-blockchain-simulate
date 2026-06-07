package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionFilter_Validate_Defaults(t *testing.T) {
	// Zero value -> defaults.
	f := TransactionFilter{}
	f.Validate()
	assert.Equal(t, 1, f.Page, "Page<1 harus jadi 1")
	assert.Equal(t, 10, f.Limit, "Limit tidak valid harus jadi 10")
	assert.Equal(t, "all", f.Type, "Type kosong harus jadi 'all'")
	assert.Equal(t, "ALL", f.Status, "Status kosong harus jadi 'ALL'")
	assert.Equal(t, "id", f.SortBy, "SortBy invalid harus jadi 'id'")
	assert.Equal(t, "DESC", f.Order, "Order invalid harus jadi 'DESC'")
}

func TestTransactionFilter_Validate_UpperLimit(t *testing.T) {
	// Limit > 100 -> default 10.
	f := TransactionFilter{Limit: 500}
	f.Validate()
	assert.Equal(t, 10, f.Limit)
}

func TestTransactionFilter_Validate_NegativePage(t *testing.T) {
	f := TransactionFilter{Page: -5, Limit: 0}
	f.Validate()
	assert.Equal(t, 1, f.Page)
	assert.Equal(t, 10, f.Limit)
}

func TestTransactionFilter_Validate_TypeNormalization(t *testing.T) {
	// Type dinormalisasi ke lowercase + trim.
	tests := []struct {
		in   string
		want string
	}{
		{"ALL", "all"},
		{"BUY", "buy"},
		{"  sell  ", "sell"},
		{"send", "send"},
		{"received", "received"},
		{"garbage", "all"},
		{"", "all"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			f := TransactionFilter{Type: tt.in}
			f.Validate()
			assert.Equal(t, tt.want, f.Type)
		})
	}
}

func TestTransactionFilter_Validate_StatusNormalization(t *testing.T) {
	// Status dinormalisasi ke uppercase + trim.
	tests := []struct {
		in   string
		want string
	}{
		{"pending", "PENDING"},
		{"CONFIRMED", "CONFIRMED"},
		{"  all  ", "ALL"},
		{"", "ALL"},
		{"rejected", "ALL"},
		{"pending,confirmed", "ALL"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			f := TransactionFilter{Status: tt.in}
			f.Validate()
			assert.Equal(t, tt.want, f.Status)
		})
	}
}

func TestTransactionFilter_Validate_OrderNormalization(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"asc", "ASC"},
		{"desc", "DESC"},
		{"ASC", "ASC"},
		{"DESC", "DESC"},
		{"foo", "DESC"},
		{"", "DESC"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			f := TransactionFilter{Order: tt.in}
			f.Validate()
			assert.Equal(t, tt.want, f.Order)
		})
	}
}

func TestTransactionFilter_Validate_ValidValuesPreserved(t *testing.T) {
	// Nilai valid harus tidak dimodifikasi.
	f := TransactionFilter{
		Page:   5,
		Limit:  25,
		Type:   "BUY",
		Status: "PENDING",
		SortBy: "amount",
		Order:  "ASC",
	}
	f.Validate()
	assert.Equal(t, 5, f.Page)
	assert.Equal(t, 25, f.Limit)
	assert.Equal(t, "buy", f.Type)
	assert.Equal(t, "PENDING", f.Status)
	assert.Equal(t, "amount", f.SortBy)
	assert.Equal(t, "ASC", f.Order)
}
