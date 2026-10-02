package worker

import (
	"context"
	"log"
	"time"
)

type OutboxPublisher struct {
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
			}
		}
	}()
}
