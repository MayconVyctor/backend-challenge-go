package database

import (
	"backend-challenge-go/internal/domain"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxWalletRepository struct {
	db *pgxpool.Pool
}

func NewPgxWalletRepository(db *pgxpool.Pool) domain.WalletRepository {
	return &pgxWalletRepository{db: db}
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
	return nil, nil
}
