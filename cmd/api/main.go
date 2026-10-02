package api

import (
	"backend-challenge-go/internal/application"
	"backend-challenge-go/internal/infrastructure/database"
	httpPresentation "backend-challenge-go/internal/presentation/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(

		fx.Provide(
			database.NewDatabasePool,
			database.NewPgxWalletRepository,
			application.NewCreateWalletUseCase,
			application.NewTransactionUseCase,
			httpPresentation.NewWalletHandler,
			httpPresentation.NewHTTPServer,
		),

		fx.Invoke(func(*echo.Echo) {}),
	)

	app.Run()
}
