package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WagerTransaction struct {
	id                    string
	providerId            string
	externalTransactionId string
	idempotencyKey        string
	walletId              string
	playerId              string
	kind                  string
	amount                Money
	status                string
	createdAt             time.Time
}

func NewWagerTransaction(providerId, externalTxId, idempotencyKey, walletId, playerId, kind string, amount Money) (WagerTransaction, error) {
	if providerId == "" || externalTxId == "" || idempotencyKey == "" || walletId == "" || playerId == "" || kind == "" {
		return WagerTransaction{}, errors.New("missing required fields for wager transaction")
	}

	return WagerTransaction{
		id:                    uuid.New().String(),
		providerId:            providerId,
		externalTransactionId: externalTxId,
		idempotencyKey:        idempotencyKey,
		walletId:              walletId,
		playerId:              playerId,
		kind:                  kind,
		amount:                amount,
		status:                "PROCESSED",
		createdAt:             time.Now().UTC(),
	}, nil
}

func (w *WagerTransaction) ID() string { return w.id }
func (w *WagerTransaction) ProviderID() string { return w.providerId }
func (w *WagerTransaction) ExternalTransactionID() string { return w.externalTransactionId }
func (w *WagerTransaction) IdempotencyKey() string { return w.idempotencyKey }
func (w *WagerTransaction) WalletID() string { return w.walletId }
func (w *WagerTransaction) PlayerID() string { return w.playerId }
func (w *WagerTransaction) Kind() string { return w.kind }
func (w *WagerTransaction) Amount() Money { return w.amount }
func (w *WagerTransaction) Status() string { return w.status }
func (w *WagerTransaction) CreatedAt() time.Time { return w.createdAt }

type WalletLedgerEntry struct {
	id            string
	walletId      string
	transactionId string
	direction     string
	amount        Money
	balanceBefore Money
	balanceAfter  Money
	createdAt     time.Time
}

func NewWalletLedgerEntry(walletId, transactionId, direction string, amount, balanceBefore, balanceAfter Money) WalletLedgerEntry {
	return WalletLedgerEntry{
		id:            uuid.New().String(),
		walletId:      walletId,
		transactionId: transactionId,
		direction:     direction,
		amount:        amount,
		balanceBefore: balanceBefore,
		balanceAfter:  balanceAfter,
		createdAt:     time.Now().UTC(),
	}
}

func (l *WalletLedgerEntry) ID() string { return l.id }
func (l *WalletLedgerEntry) WalletID() string { return l.walletId }
func (l *WalletLedgerEntry) TransactionID() string { return l.transactionId }
func (l *WalletLedgerEntry) Direction() string { return l.direction }
func (l *WalletLedgerEntry) Amount() Money { return l.amount }
func (l *WalletLedgerEntry) BalanceBefore() Money { return l.balanceBefore }
func (l *WalletLedgerEntry) BalanceAfter() Money { return l.balanceAfter }
func (l *WalletLedgerEntry) CreatedAt() time.Time { return l.createdAt }
