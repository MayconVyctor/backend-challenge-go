package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func NewDatabasePool(lc fx.Lifecycle) *pgxpool.Pool {

	dbUrl := "postgres://jungle:jungle123@localhost:5432/walletdb?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			fmt.Println("Closing db conection")
			pool.Close()
			return nil
		},
	})

	return pool
}
