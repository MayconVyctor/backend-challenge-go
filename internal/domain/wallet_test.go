package domain

import (
	"testing"
)

func TestWallet_Debit_InsufficientFunds(t *testing.T) {

	initialBalance, _ := NewMoneyFromString("10.00", "BRL")
	wallet, _ := NewWallet("player-1", initialBalance)

	debitAmount, _ := NewMoneyFromString("15.00", "BRL")

	err := wallet.Debit(debitAmount)

	if err == nil {
		t.Errorf("I was expecting an insufficient funds error, but the transaction was approved")
	}

	if wallet.balance.amount != 1000 {
		t.Errorf("expected balance amount to be 1000, got %d", wallet.balance.amount)
	}

	if wallet.version != 1 {
		t.Errorf("expected version to be 1, got %d", wallet.version)
	}
}
