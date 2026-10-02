package worker

import (
	"context"
	"log"
	"time"
)

type OutboxPublisher struct {
	// in a real app, this would receive the DB connection and an Event Publisher interface
}

func NewOutboxPublisher() *OutboxPublisher {
	return &OutboxPublisher{}
}

func (p *OutboxPublisher) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("OutboxPublisher shutting down...")
				return
			case <-ticker.C:
				// 1. SELECT * FROM outbox WHERE status = 'PENDING' FOR UPDATE SKIP LOCKED
				// 2. Publish to SNS/SQS/Kafka
				// 3. UPDATE outbox SET status = 'PUBLISHED', published_at = NOW() WHERE event_id = $1
			}
		}
	}()
}
