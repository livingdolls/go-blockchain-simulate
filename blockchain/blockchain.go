package blockchain

import (
	"fmt"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/block"
	"github.com/livingdolls/go-blockchain-simulate/signature"
	"github.com/livingdolls/go-blockchain-simulate/transaction"
	"github.com/livingdolls/go-blockchain-simulate/wallet"
)

type Blockchain struct {
	Chain   []block.Block
	Mempool []transaction.Transaction
	Wallets map[string]wallet.Wallet
}

func NewBlockchain() *Blockchain {
	genesis := block.Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []transaction.Transaction{},
		PrevHash:     "0",
	}

	genesis.Hash = block.HashBlock(genesis)

	return &Blockchain{
		Chain:   []block.Block{genesis},
		Mempool: []transaction.Transaction{},
		Wallets: make(map[string]wallet.Wallet),
	}
}

func (bc *Blockchain) RegisterWallet(w wallet.Wallet) {
	fmt.Println("Registering wallet:", w.Address)
	bc.Wallets[w.Address] = w
}

func (bc *Blockchain) AddSignedTransaction(tx transaction.Transaction) error {

	fmt.Println("=== Debug Transaction ===")
	fmt.Println("Transaction From:", tx.From)
	fmt.Println("Registered Wallets:")
	for addr := range bc.Wallets {
		fmt.Println("  -", addr)
	}
	fmt.Println("========================")

	// Verify sender wallet exists
	w, ok := bc.Wallets[tx.From]
	if !ok {
		return fmt.Errorf("wallet sender not found")
	}

	// verify transaction signature
	if !signature.VerifySignature(w.PrivateKey, tx.Message, tx.Signature) {
		return fmt.Errorf("invalid transaction signature")
	}

	// Add transaction to mempool
	bc.Mempool = append(bc.Mempool, tx)
	return nil
}

func (bc *Blockchain) MineBlock() block.Block {
	prev := bc.Chain[len(bc.Chain)-1]

	newBlock := block.Block{
		Index:        len(bc.Chain),
		Timestamp:    time.Now().Unix(),
		Transactions: bc.Mempool,
		PrevHash:     prev.PrevHash,
	}

	block.HashBlock(newBlock)
	bc.Chain = append(bc.Chain, newBlock)
	bc.Mempool = []transaction.Transaction{}
	return newBlock
}

func (bc *Blockchain) GetBalance(addr string) int64 {
	var balance int64 = 0

	for _, block := range bc.Chain {
		for _, tx := range block.Transactions {
			if tx.From == addr {
				balance -= tx.Amount
			}

			if tx.To == addr {
				balance += tx.Amount
			}
		}
	}

	return balance
}
