package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func NewDatabasePool(lc fx.Lifecycle) *pgxpool.Pool {

	dbUrl := os.Getenv("DB_URL")

	if err := RunMigrations(dbUrl); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failure: %v\n", err)
		os.Exit(1)
	}

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
