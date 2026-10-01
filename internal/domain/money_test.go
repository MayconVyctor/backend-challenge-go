package domain

import (
	"testing"
)

func TestNewMoneyFromString_Success(t *testing.T) {

	amountStr := "25.00"
	currency := "BRL"

	money, err := NewMoneyFromString(amountStr, currency)

	if err != nil {
		t.Errorf("I wasnt expecting an error, but I received: %v", err)
	}

	if money.amount != 2500 {
		t.Errorf("was expecting amount 2500, but I received %d", money.amount)
	}

	if money.currency != "BRL" {
		t.Errorf("was expecting BRL, but I received %s", money.currency)
	}
}
