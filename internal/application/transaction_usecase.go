package application

import (
	"backend-challenge-go/internal/domain"
	"context"
	"errors"
)

type ProcessTransactionInput struct {
	IdempotencyKey        string
	ProviderID            string
	ExternalTransactionID string
	PlayerID              string
	Amount                string
	Currency              string
	Kind                  string
}

type TransactionUseCase struct {
	repo domain.WalletRepository
}

func NewTransactionUseCase(repo domain.WalletRepository) *TransactionUseCase {
	return &TransactionUseCase{repo: repo}
}

func (uc *TransactionUseCase) Execute(ctx context.Context, input ProcessTransactionInput) (*domain.Wallet, error) {

	exists, err := uc.repo.HasIdempotencyKey(ctx, input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if exists {
		return uc.repo.FindByID(ctx, input.PlayerID)
	}

	money, err := domain.NewMoneyFromString(input.Amount, input.Currency)
	if err != nil {
		return nil, err
	}

	wallet, err := uc.repo.FindByID(ctx, input.PlayerID)
	if err != nil {
		return nil, err
	}

	if input.Kind == "BET" {
		err = wallet.Debit(money)
		if err != nil {
			return nil, err
		}
	} else if input.Kind == "WIN" {
		err = wallet.Credit(money)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("invalid transaction kind")
	}

	err = uc.repo.Save(ctx, wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}
