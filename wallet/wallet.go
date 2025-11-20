package wallet

import "github.com/livingdolls/go-blockchain-simulate/utils"

type Wallet struct {
	Address    string
	PrivateKey string
}

func NewWallet() Wallet {
	return Wallet{
		Address:    "0x" + utils.RandomHex(20),
		PrivateKey: utils.RandomHex(32),
	}
}
