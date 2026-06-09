package models

type UserRegister struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Address   string `json:"address" binding:"required,eth_addr"`
	PublicKey string `json:"public_key" binding:"required,hex"`
}
