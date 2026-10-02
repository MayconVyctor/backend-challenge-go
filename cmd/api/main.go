package api

import (
	"context"
	"fmt"

	"backend-challenge-go/internal/application"
	"backend-challenge-go/internal/infrastructure/database"
	httpPresentation "backend-challenge-go/internal/presentation/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func NewHTTPServer(lc fx.Lifecycle, handler *httpPresentation.WalletHandler) *echo.Echo {

	e := echo.New()

	e.POST("/wallets", handler.CreateWallet)
	e.POST("/wagering/transactions", handler.ProcessTransaction)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("Initialize server HTTP on port 8080")
			go e.Start(":8080")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Turn of server HTTP safely")
			return e.Shutdown(ctx)
		},
	})
	return e
}

func main() {
	app := fx.New(

		fx.Provide(
			database.NewDatabasePool,
			database.NewPgxWalletRepository,
			application.NewCreateWalletUseCase,
			application.NewTransactionUseCase,
			httpPresentation.NewWalletHandler,
			NewHTTPServer,
		),

		fx.Invoke(func(*echo.Echo) {}),
	)

	app.Run()
}
