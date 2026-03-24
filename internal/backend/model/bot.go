package model

import (
	"context"
	"time"

	"mcd-bot/internal/backend/enum"
)

// Bot represents a cooking bot that processes orders.
type Bot struct {
	ID           int
	Status       enum.BotStatus
	CurrentOrder *Order
	cancelFunc   context.CancelFunc
}

// NewBot creates a new bot with the given ID, defaulting to Idle status.
func NewBot(id int) *Bot {
	return &Bot{
		ID:     id,
		Status: enum.Idle,
	}
}

// Process starts processing the given order in a goroutine.
// It calls onComplete when the order finishes after processingTime.
func (b *Bot) Process(order *Order, processingTime time.Duration, onComplete func(bot *Bot, order *Order)) {
	b.Status = enum.BotActive
	b.CurrentOrder = order
	order.Status = enum.Processing

	ctx, cancel := context.WithCancel(context.Background())
	b.cancelFunc = cancel

	go func() {
		select {
		case <-time.After(processingTime):
			order.Status = enum.Complete
			b.Status = enum.Idle
			b.CurrentOrder = nil
			b.cancelFunc = nil
			onComplete(b, order)
		case <-ctx.Done():
			// Processing was cancelled (bot removed)
			return
		}
	}()
}

// Cancel stops the bot's current processing and returns the in-progress order.
// Returns nil if the bot was idle.
func (b *Bot) Cancel() *Order {
	if b.cancelFunc != nil {
		b.cancelFunc()
		b.cancelFunc = nil
	}

	order := b.CurrentOrder
	if order != nil {
		order.Status = enum.Pending
		b.CurrentOrder = nil
	}

	b.Status = enum.Idle
	return order
}
