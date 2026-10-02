package application

import (
	"backend-challenge-go/internal/domain"
	"context"
)

type CreateWalletInput struct {
	PlayerID       string
	InitialBalance string
	Currency       string
}

type CreateWalletUseCase struct {
	repo domain.WalletRepository
}

func NewCreateWalletUseCase(repo domain.WalletRepository) *CreateWalletUseCase {
	return &CreateWalletUseCase{repo: repo}
}

func (uc *CreateWalletUseCase) Execute(ctx context.Context, input CreateWalletInput) (*domain.Wallet, error) {

	money, err := domain.NewMoneyFromString(input.InitialBalance, input.Currency)
	if err != nil {
		return nil, err
	}

	wallet, err := domain.NewWallet(input.PlayerID, money)
	if err != nil {
		return nil, err
	}

	err = uc.repo.Save(ctx, &wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}
