package models

type User struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	Address   string `db:"address"`
	PublicKey string `db:"public_key"`
}

type UserRegisterResponse struct {
	Address    string  `json:"address"`
	Username   string  `json:"username"`
	YTEBalance float64 `json:"yt_balance,omitempty"`
	USDBalance float64 `json:"usd_balance,omitempty"`
	Token      string  `json:"token,omitempty"`
}

type UserLoginResponse struct {
	ID         int     `json:"id"`
	Address    string  `json:"address"`
	Username   string  `json:"username"`
	YTEBalance float64 `json:"yt_balance,omitempty"`
	USDBalance float64 `json:"usd_balance,omitempty"`
	PublicKey  string  `json:"public_key,omitempty"`
}

type UserWithBalance struct {
	ID         int     `db:"id"`
	Name       string  `db:"name"`
	Address    string  `db:"address"`
	PublicKey  string  `db:"public_key"`
	YTEBalance float64 `db:"yte_balance"`
	USDBalance float64 `db:"usd_balance"`
}
