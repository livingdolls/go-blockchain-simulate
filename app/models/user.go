package models

type User struct {
	ID         int     `db:"id"`
	Name       string  `db:"name"`
	Address    string  `db:"address"`
	PublicKey  string  `db:"public_key"`
	PrivateKey string  `db:"private_key"`
	Balance    float64 `db:"balance"`
}
