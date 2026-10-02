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
	WalletID              string
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

type ProcessTransactionOutput struct {
	TransactionID    string
	Status           string
	Balance          string
	Currency         string
	IdempotentReplay bool
}

func (uc *TransactionUseCase) Execute(ctx context.Context, input ProcessTransactionInput) (ProcessTransactionOutput, error) {

	money, err := domain.NewMoneyFromString(input.Amount, input.Currency)
	if err != nil {
		return ProcessTransactionOutput{}, err
	}

	var output ProcessTransactionOutput

	err = uc.repo.RunInTransaction(ctx, func(txCtx context.Context) error {

		exists, err := uc.repo.HasIdempotencyKey(txCtx, input.IdempotencyKey)
		if err != nil {
			return err
		}
		if exists {
			// For now, if idempotency key exists, we just fetch the wallet to return its balance.
			// Ideally, we should fetch the exact transaction and verify the payload hash as per the challenge.
			wallet, err := uc.repo.FindByID(txCtx, input.WalletID)
			if err != nil {
				return err
			}
			
			// Format balance to string (e.g., 25.00)
			balanceStr := formatBalance(wallet.Balance().Amount())
			
			output = ProcessTransactionOutput{
				TransactionID:    "existing-tx-id", // Placeholder
				Status:           "PROCESSED",
				Balance:          balanceStr,
				Currency:         wallet.Currency(),
				IdempotentReplay: true,
			}
			return nil
		}

		wallet, err := uc.repo.FindByID(txCtx, input.WalletID)
		if err != nil {
			return err
		}

		balanceBefore := wallet.Balance()

		var direction string
		if input.Kind == "BET" {
			err = wallet.Debit(money)
			direction = "DEBIT"
		} else if input.Kind == "WIN" {
			err = wallet.Credit(money)
			direction = "CREDIT"
		} else {
			return errors.New("invalid transaction kind")
		}

		if err != nil {
			return err
		}

		err = uc.repo.Save(txCtx, wallet)
		if err != nil {
			return err
		}

		wagerTx, err := domain.NewWagerTransaction(
			input.ProviderID,
			input.ExternalTransactionID,
			input.IdempotencyKey,
			input.WalletID,
			input.PlayerID,
			input.Kind,
			money,
		)
		if err != nil {
			return err
		}

		err = uc.repo.SaveWagerTransaction(txCtx, &wagerTx)
		if err != nil {
			return err
		}

		ledgerEntry := domain.NewWalletLedgerEntry(
			wallet.ID(),
			wagerTx.ID(),
			direction,
			money,
			balanceBefore,
			wallet.Balance(),
		)

		err = uc.repo.SaveLedgerEntry(txCtx, &ledgerEntry)
		if err != nil {
			return err
		}

		balanceStr := formatBalance(wallet.Balance().Amount())

		output = ProcessTransactionOutput{
			TransactionID:    wagerTx.ID(),
			Status:           wagerTx.Status(),
			Balance:          balanceStr,
			Currency:         wallet.Currency(),
			IdempotentReplay: false,
		}

		return nil
	})

	if err != nil {
		return ProcessTransactionOutput{}, err
	}

	return output, nil
}

import "fmt"

func formatBalance(amount int64) string {
	return fmt.Sprintf("%d.%02d", amount/100, amount%100)
}
