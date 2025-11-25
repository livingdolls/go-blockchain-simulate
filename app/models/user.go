package models

type User struct {
	ID        int     `db:"id"`
	Name      string  `db:"name"`
	Address   string  `db:"address"`
	PublicKey string  `db:"public_key"`
	Balance   float64 `db:"balance"`
}

type UserRegisterResponse struct {
	Mnemonic   string `json:"mnemonic"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Username   string `json:"username"`
}
