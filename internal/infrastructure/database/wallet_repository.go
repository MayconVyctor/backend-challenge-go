package database

import (
	"backend-challenge-go/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxWalletRepository struct {
	db *pgxpool.Pool
}

type QueryEngine interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func NewPgxWalletRepository(db *pgxpool.Pool) domain.WalletRepository {
	return &pgxWalletRepository{db: db}
}

func (r *pgxWalletRepository) getQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value("tx").(pgx.Tx)
	if ok {
		return tx
	}
	return r.db
}

func (r *pgxWalletRepository) RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, "tx", tx)

	if err := fn(txCtx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (r *pgxWalletRepository) Save(ctx context.Context, w *domain.Wallet) error {
	query := `
		INSERT INTO wallets (id, player_id, currency, balance, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		w.ID(),
		w.PlayerID(),
		w.Currency(),
		w.Balance().Amount(),
		w.Version(),
		w.CreatedAt(),
		w.UpdatedAt(),
	)

	return err
}

func (r *pgxWalletRepository) FindByID(ctx context.Context, id string) (*domain.Wallet, error) {

	query := `
		SELECT id, player_id, currency, balance, version, created_at, updated_at
		FROM wallets
		WHERE id = $1
		FOR UPDATE
	`

	var (
		wID        string
		wPlayerID  string
		wCurrency  string
		wBalance   int64
		wVersion   int
		wCreatedAt time.Time
		wUpdatedAt time.Time
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&wID,
		&wPlayerID,
		&wCurrency,
		&wBalance,
		&wVersion,
		&wCreatedAt,
		&wUpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}

	wallet := domain.RestoreWallet(
		wID,
		wPlayerID,
		wCurrency,
		wBalance,
		wVersion,
		wCreatedAt,
		wUpdatedAt,
	)

	return wallet, nil
}

func (r *pgxWalletRepository) HasIdempotencyKey(ctx context.Context, key string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM wager_transactions WHERE idempotency_key = $1)`

	err := r.db.QueryRow(ctx, query, key).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
