package application

import (
	"backend-challenge-go/internal/domain"
	"context"
	"errors"
	"fmt"
)

type ProcessTransactionInput struct {
	MessageID                      string
	ConsumerName                   string
	IdempotencyKey                 string
	ProviderID                     string
	ExternalTransactionID          string
	ReferenceExternalTransactionID string
	PlayerID                       string
	WalletID                       string
	Amount                         string
	Currency                       string
	Kind                           string
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

		if input.MessageID != "" && input.ConsumerName != "" {
			acquired, err := uc.repo.TryAcquireInboxMessage(txCtx, input.ConsumerName, input.MessageID)
			if err != nil {
				return err
			}
			if !acquired {
				return nil
			}
		}

		exists, err := uc.repo.HasIdempotencyKey(txCtx, input.IdempotencyKey)
		if err != nil {
			return err
		}
		if exists {
			wallet, err := uc.repo.FindByID(txCtx, input.WalletID)
			if err != nil {
				return err
			}
			balanceStr := formatBalance(wallet.Balance().Amount())
			output = ProcessTransactionOutput{
				TransactionID:    "existing-tx-id",
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
		} else if input.Kind == "WIN" || input.Kind == "REFUND" || input.Kind == "ROLLBACK" {
			if input.Kind == "REFUND" || input.Kind == "ROLLBACK" {
				if input.ReferenceExternalTransactionID == "" {
					return errors.New("missing reference transaction for reversal")
				}
				_, err := uc.repo.FindWagerByExternalID(txCtx, input.ProviderID, input.ReferenceExternalTransactionID)
				if err != nil {
					return errors.New("reference transaction not found")
				}
			}
			err = wallet.Credit(money)
			direction = "CREDIT"
		} else if input.Kind == "LOSS" {
			if money.Amount() != 0 {
				return errors.New("LOSS must have amount 0")
			}
			direction = "NONE"
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

		if direction != "NONE" {
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
		}

		balanceStr := formatBalance(wallet.Balance().Amount())

		if direction != "NONE" {
			walletEvent := domain.NewWalletBalanceChanged(
				wallet.ID(),
				wagerTx.ID(),
				direction,
				formatBalance(money.Amount()),
				money.Currency(),
				formatBalance(balanceBefore.Amount()),
				balanceStr,
				wallet.Version(),
			)
			walletOutbox := domain.NewOutboxEntry(walletEvent)
			err = uc.repo.SaveOutboxEntry(txCtx, &walletOutbox)
			if err != nil {
				return err
			}
		}

		wagerEvent := domain.NewWagerTransactionProcessed(wagerTx.ID())
		wagerOutbox := domain.NewOutboxEntry(wagerEvent)
		err = uc.repo.SaveOutboxEntry(txCtx, &wagerOutbox)
		if err != nil {
			return err
		}

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

func formatBalance(amount int64) string {
	return fmt.Sprintf("%d.%02d", amount/100, amount%100)
}
