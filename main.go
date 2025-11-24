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

	balanceService := services.NewBalanceService(userRepo, txRepo)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	blockRepo := repository.NewBlockRepository(db.GetDB())
	blockService := services.NewBlockService(blockRepo, txRepo, userRepo, ledgerRepo)
	blockHandler := handler.NewBlockHandler(blockService)

	rewardService := services.NewRewardHandler(blockRepo)
	rewardHandler := handler.NewRewardHandler(rewardService, blockService)

	r := gin.Default()

	r.POST("/register", userHandler.Register)
	r.POST("/send", transactionHandler.Send)
	r.GET("/transaction/:id", transactionHandler.GetTransaction)
	r.GET("/balance/:address", balanceHandler.GetBalance)
	r.GET("/wallet/:address", balanceHandler.GetWalletBalance)
	r.POST("/generate-block", blockHandler.GenerateBlock)
	r.GET("/blocks", blockHandler.GetBlocks)
	r.GET("/blocks/:id", blockHandler.GetBlockByID)
	r.GET("/blocks/detail/:number", blockHandler.GetBlockByBlockNumber)
	r.GET("/blocks/integrity", blockHandler.CheckBlockchainIntegrity)
	r.GET("/reward/schedule/:number", rewardHandler.GetRewardSchedule)
	r.GET("/reward/block/:number", rewardHandler.GetBlockReward)
	r.GET("/reward/info", rewardHandler.GetRewardInfo)

	r.Run(":3010")
}
