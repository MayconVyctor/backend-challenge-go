package http

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func NewHTTPServer(lc fx.Lifecycle, handler *WalletHandler) *echo.Echo {

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
