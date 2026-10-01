package domain

import (
	"context"
	"errors"
	"time"
)

type Wallet struct {
	id        string
	playerId  string
	currency  string
	balance   Money
	version   int
	createdAt time.Time
	updatedAt time.Time
}

func NewWallet(playerId string, initialBalance Money) (Wallet, error) {

	if playerId == "" {
		return Wallet{}, errors.New("playerId cannot be blank")
	}

	return Wallet{
		id:        "temp-id-123",
		playerId:  playerId,
		currency:  initialBalance.currency,
		balance:   initialBalance,
		version:   1,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
	}, nil
}

func (w *Wallet) Debit(amount Money) error {

	newBalance, err := w.balance.Subtract(amount)
	if err != nil {
		return err
	}

	if newBalance.amount < 0 {
		return errors.New("insufficient funds")
	}

	w.balance = newBalance
	w.version++
	w.updatedAt = time.Now().UTC()

	return nil
}

func (w *Wallet) Credit(amount Money) error {

	newBalance, err := w.balance.Add(amount)
	if err != nil {
		return err
	}

	w.balance = newBalance
	w.version++
	w.updatedAt = time.Now().UTC()

	return nil
}

type WalletRepository interface {
	FindByID(ctx context.Context, id string) (*Wallet, error)
	Save(ctz context.Context, w *Wallet) error
}
