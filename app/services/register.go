package services

import (
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type RegisterService interface {
	Register(name string) (models.UserRegisterResponse, error)
}

type registerService struct {
	repo repository.UserRepository
}

func NewRegisterService(repo repository.UserRepository) RegisterService {
	return &registerService{repo: repo}
}

// Registr implements RegisterService.
func (r *registerService) Register(name string) (models.UserRegisterResponse, error) {
	// create mnemonic
	mnemonic, err := utils.GenerateMnemonic()
	if err != nil {
		return models.UserRegisterResponse{}, err
	}

	// derive wallet
	_, privHex, pubHex, addr, err := utils.GenerateWalletFromMnemonic(mnemonic, "")
	if err != nil {
		return models.UserRegisterResponse{}, err
	}

	// save to db
	user := models.User{
		Name:      name,
		Address:   addr,
		PublicKey: pubHex,
		Balance:   1000,
	}

	err = r.repo.Create(user)
	if err != nil {
		return models.UserRegisterResponse{}, err
	}

	userResponse := models.UserRegisterResponse{
		Username:   name,
		Mnemonic:   mnemonic,
		Address:    addr,
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}

	return userResponse, nil
}
