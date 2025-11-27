package models

type UserRegister struct {
	Username  string `json:"username"`
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
}
