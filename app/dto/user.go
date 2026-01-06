package dto

type DTOUserWithBalance struct {
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	YTEBalance float64 `json:"yte_balance"`
	USDBalance float64 `json:"usd_balance"`
}
