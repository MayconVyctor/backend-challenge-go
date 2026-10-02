package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"backend-challenge-go/internal/application"
)

type SQSMessage struct {
	MessageID string
	Body      string
}

type SQSClient interface {
	ReceiveMessages(ctx context.Context) ([]SQSMessage, error)
	DeleteMessage(ctx context.Context, receiptHandle string) error
}

type SQSConsumer struct {
	client SQSClient
	uc     *application.TransactionUseCase
}

func NewSQSConsumer(client SQSClient, uc *application.TransactionUseCase) *SQSConsumer {
	return &SQSConsumer{
		client: client,
		uc:     uc,
	}
}

func (c *SQSConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("SQSConsumer shutting down...")
				return
			default:
				messages, err := c.client.ReceiveMessages(ctx)
				if err != nil {
					log.Printf("Error receiving SQS messages: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}

				for _, msg := range messages {
					c.processMessage(ctx, msg)
				}
			}
		}
	}()
}

type sqsPayload struct {
	Data struct {
		ProviderID            string `json:"providerId"`
		ExternalTransactionID string `json:"externalTransactionId"`
		IdempotencyKey        string `json:"idempotencyKey"`
		PlayerID              string `json:"playerId"`
		WalletID              string `json:"walletId"`
		Kind                  string `json:"kind"`
		Money                 struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"money"`
	} `json:"data"`
}

func (c *SQSConsumer) processMessage(ctx context.Context, msg SQSMessage) {
	var payload sqsPayload
	if err := json.Unmarshal([]byte(msg.Body), &payload); err != nil {
		log.Printf("Error parsing SQS message body: %v", err)
		return
	}

	input := application.ProcessTransactionInput{
		MessageID:             msg.MessageID,
		ConsumerName:          "wager-sqs-consumer",
		IdempotencyKey:        payload.Data.IdempotencyKey,
		ProviderID:            payload.Data.ProviderID,
		ExternalTransactionID: payload.Data.ExternalTransactionID,
		PlayerID:              payload.Data.PlayerID,
		WalletID:              payload.Data.WalletID,
		Kind:                  payload.Data.Kind,
		Amount:                payload.Data.Money.Amount,
		Currency:              payload.Data.Money.Currency,
	}

	_, err := c.uc.Execute(ctx, input)
	if err != nil {
		log.Printf("Error processing message %s: %v", msg.MessageID, err)
		return
	}

	fmt.Printf("Successfully processed message %s\n", msg.MessageID)
}
