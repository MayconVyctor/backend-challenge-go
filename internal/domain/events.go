package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event interface {
	EventID() string
	EventType() string
	AggregateID() string
	Payload() []byte
}

type OutboxEntry struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     []byte
	Status      string
	CreatedAt   time.Time
}

func NewOutboxEntry(event Event) OutboxEntry {
	return OutboxEntry{
		EventID:     event.EventID(),
		EventType:   event.EventType(),
		AggregateID: event.AggregateID(),
		Payload:     event.Payload(),
		Status:      "PENDING",
		CreatedAt:   time.Time{}, // handled by DB default or set explicitly
	}
}

type WalletBalanceChanged struct {
	ID            string
	Type          string
	WalletID      string
	TransactionID string
	Direction     string
	MoneyAmount   string
	MoneyCurrency string
	BalanceBefore string
	BalanceAfter  string
	WalletVersion int
	OccurredAt    time.Time
}

func NewWalletBalanceChanged(walletId, transactionId, direction, moneyAmount, moneyCurrency, balanceBefore, balanceAfter string, version int) WalletBalanceChanged {
	return WalletBalanceChanged{
		ID:            uuid.New().String(),
		Type:          "WalletBalanceChanged",
		WalletID:      walletId,
		TransactionID: transactionId,
		Direction:     direction,
		MoneyAmount:   moneyAmount,
		MoneyCurrency: moneyCurrency,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		WalletVersion: version,
		OccurredAt:    time.Now().UTC(),
	}
}

func (e WalletBalanceChanged) EventID() string { return e.ID }
func (e WalletBalanceChanged) EventType() string { return e.Type }
func (e WalletBalanceChanged) AggregateID() string { return e.WalletID }
func (e WalletBalanceChanged) Payload() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}

type WagerTransactionProcessed struct {
	ID            string
	Type          string
	TransactionID string
	OccurredAt    time.Time
}

func NewWagerTransactionProcessed(transactionId string) WagerTransactionProcessed {
	return WagerTransactionProcessed{
		ID:            uuid.New().String(),
		Type:          "WagerTransactionProcessed",
		TransactionID: transactionId,
		OccurredAt:    time.Now().UTC(),
	}
}

func (e WagerTransactionProcessed) EventID() string { return e.ID }
func (e WagerTransactionProcessed) EventType() string { return e.Type }
func (e WagerTransactionProcessed) AggregateID() string { return e.TransactionID }
func (e WagerTransactionProcessed) Payload() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}
