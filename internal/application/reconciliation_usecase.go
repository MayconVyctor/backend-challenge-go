package application

import (
	"context"
	"fmt"
	"backend-challenge-go/internal/domain"
)

type ReconciliationOutput struct {
	WalletID          string `json:"walletId"`
	StoredBalance     map[string]string `json:"storedBalance"`
	CalculatedBalance map[string]string `json:"calculatedBalance"`
	Difference        map[string]string `json:"difference"`
	Consistent        bool `json:"consistent"`
	CheckedEntries    int `json:"checkedEntries"`
}

type ReconciliationUseCase struct {
	repo domain.WalletRepository
}

func NewReconciliationUseCase(repo domain.WalletRepository) *ReconciliationUseCase {
	return &ReconciliationUseCase{repo: repo}
}

func (uc *ReconciliationUseCase) Execute(ctx context.Context, walletID string) (ReconciliationOutput, error) {
	var output ReconciliationOutput
	
	err := uc.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
		wallet, err := uc.repo.FindByID(txCtx, walletID)
		if err != nil {
			return err
		}

		calcBalance, count, err := uc.repo.GetLedgerBalanceAndCount(txCtx, walletID)
		if err != nil {
			return err
		}

		diff := wallet.Balance().Amount() - calcBalance
		isConsistent := diff == 0

		output = ReconciliationOutput{
			WalletID: walletID,
			StoredBalance: map[string]string{"amount": formatBalance(wallet.Balance().Amount()), "currency": wallet.Currency()},
			CalculatedBalance: map[string]string{"amount": formatBalance(calcBalance), "currency": wallet.Currency()},
			Difference: map[string]string{"amount": formatBalance(diff), "currency": wallet.Currency()},
			Consistent: isConsistent,
			CheckedEntries: count,
		}
		return nil
	})

	return output, err
}
