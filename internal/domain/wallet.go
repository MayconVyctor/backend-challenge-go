package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
		id:        uuid.New().String(),
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
	Save(ctx context.Context, w *Wallet) error
	HasIdempotencyKey(ctx context.Context, key string) (bool, error)
	RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	SaveWagerTransaction(ctx context.Context, tx *WagerTransaction) error
	SaveLedgerEntry(ctx context.Context, entry *WalletLedgerEntry) error
}

func RestoreWallet(id string, playerId string, currency string, balanceAmount int64, version int, createdAt time.Time, updatedAt time.Time) *Wallet {

	restoredBalance := RestoreMoney(balanceAmount, currency)

	return &Wallet{
		id:        id,
		playerId:  playerId,
		currency:  currency,
		balance:   restoredBalance,
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (w *Wallet) ID() string {
	return w.id
}

func (w *Wallet) PlayerID() string {
	return w.playerId
}

func (w *Wallet) Currency() string {
	return w.currency
}

func (w *Wallet) Balance() Money {
	return w.balance
}

func (w *Wallet) Version() int {
	return w.version
}

func (w *Wallet) CreatedAt() time.Time {
	return w.createdAt
}

func (w *Wallet) UpdatedAt() time.Time {
	return w.updatedAt
}
