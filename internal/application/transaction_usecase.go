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

	var finalWallet *domain.Wallet

	money, err := domain.NewMoneyFromString(input.Amount, input.Currency)
	if err != nil {
		return nil, err
	}

	err = uc.repo.RunInTransaction(ctx, func(txCtx context.Context) error {

		exists, err := uc.repo.HasIdempotencyKey(txCtx, input.IdempotencyKey)
		if err != nil {
			return err
		}
		if exists {
			finalWallet, err = uc.repo.FindByID(txCtx, input.PlayerID)
			return err
		}

		wallet, err := uc.repo.FindByID(txCtx, input.PlayerID)
		if err != nil {
			return err
		}

		if input.Kind == "BET" {
			err = wallet.Debit(money)
			if err != nil {
				return err
			}
		} else if input.Kind == "WIN" {
			err = wallet.Credit(money)
			if err != nil {
				return err
			}
		} else {
			return errors.New("invalid transaction kind")
		}

		err = uc.repo.Save(txCtx, wallet)
		if err != nil {
			return err
		}

		finalWallet = wallet

		return nil
	})

	if err != nil {
		return nil, err
	}

	return finalWallet, nil
}
