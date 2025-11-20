package main

import (
	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/handler"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/database"
)

func main() {

	// initialize database
	db, err := database.NewDBConn()

	if err != nil {
		panic(err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.GetDB())
	userService := services.NewRegisterService(userRepo)
	userHandler := handler.NewRegisterHandler(userService)

	txRepo := repository.NewTransactionRepository(db.GetDB())
	ledgerRepo := repository.NewLedgerRepository(db.GetDB())

	transactionService := services.NewTransactionService(userRepo, txRepo, ledgerRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	balanceService := services.NewBalanceService(userRepo)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	blockRepo := repository.NewBlockRepository(db.GetDB())
	blockService := services.NewBlockService(blockRepo, txRepo)
	blockHandler := handler.NewBlockHandler(blockService)

	r := gin.Default()

	r.POST("/register", userHandler.Register)
	r.POST("/send", transactionHandler.Send)
	r.GET("/balance/:address", balanceHandler.GetBalance)
	r.POST("/generate-block", blockHandler.GenerateBlock)
	r.GET("/blocks", blockHandler.GetBlocks)
	r.GET("/blocks/:id", blockHandler.GetBlockByID)
	r.GET("/blocks/integrity", blockHandler.CheckBlockchainIntegrity)

	r.Run(":3010")

	// initialize blockchain
	// bc := blockchain.NewBlockchain()

	// //create wallet

	// w1 := wallet.NewWallet()
	// w2 := wallet.NewWallet()

	// bc.RegisterWallet(w1)
	// bc.RegisterWallet(w2)

	// // output wallet details

	// fmt.Println("Wallet 1 Address:", w1.Address)
	// fmt.Println("Wallet 1 Private Key:", w1.PrivateKey)

	// fmt.Println("Wallet 2 Address:", w2.Address)
	// fmt.Println("Wallet 2 Private Key:", w2.PrivateKey)

	// // create transaction
	// amount := int64(100)
	// message := signature.CreateMessage(w1.Address, w2.Address, amount)
	// signatureStr := signature.Sign(w1.PrivateKey, message)

	// tx := transaction.Transaction{
	// 	From:      w1.Address,
	// 	To:        w2.Address,
	// 	Amount:    amount,
	// 	Message:   message,
	// 	Signature: signatureStr,
	// }

	// // output transaction details
	// if err := bc.AddSignedTransaction(tx); err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Transaction added to mempool:")
	// fmt.Println(" From:", tx.From)
	// fmt.Println(" To:", tx.To)
	// fmt.Println(" Amount:", tx.Amount)
	// fmt.Println(" Message:", tx.Message)
	// fmt.Println(" Signature:", tx.Signature)

	// // mine block
	// block := bc.MineBlock()

	// fmt.Println("Mined new block:")
	// fmt.Println(" Index:", block.Index)
	// fmt.Println(" Timestamp:", block.Timestamp)
	// fmt.Println(" Previous Hash:", block.PrevHash)
	// fmt.Println(" Hash:", block.Hash)
	// fmt.Println(" Transactions:", block.Transactions)

	// // check balances
	// balance1 := bc.GetBalance(w1.Address)
	// balance2 := bc.GetBalance(w2.Address)

	// fmt.Println("Balance of Wallet 1:", balance1)
	// fmt.Println("Balance of Wallet 2:", balance2)
}
