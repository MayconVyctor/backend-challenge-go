package main

import (
	"backend-challenge-go/internal/application"
	"backend-challenge-go/internal/infrastructure/database"
	"backend-challenge-go/internal/infrastructure/worker"
	httpPresentation "backend-challenge-go/internal/presentation/http"
	"context"

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
			application.NewReconciliationUseCase,
			httpPresentation.NewWalletHandler,
			httpPresentation.NewHTTPServer,
			
			func() worker.SQSClient { return nil },
			worker.NewSQSConsumer,
			worker.NewOutboxPublisher,
		),

		fx.Invoke(func(lc fx.Lifecycle, e *echo.Echo, sqsConsumer *worker.SQSConsumer, outboxPublisher *worker.OutboxPublisher) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go e.Start(":8080")
					sqsConsumer.Start(ctx)
					outboxPublisher.Start(ctx)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return e.Shutdown(ctx)
				},
			})
		}),
	)

	app.Run()
}
