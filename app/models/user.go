package models

type User struct {
	ID        int     `db:"id"`
	Name      string  `db:"name"`
	Address   string  `db:"address"`
	PublicKey string  `db:"public_key"`
	Balance   float64 `db:"balance"`
}

type UserRegisterResponse struct {
	Address  string  `json:"address"`
	Username string  `json:"username"`
	Balance  float64 `json:"balance,omitempty"`
	Token    string  `json:"token,omitempty"`
}
