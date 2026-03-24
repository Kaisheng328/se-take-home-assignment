package model

import "mcd-bot/internal/backend/enum"

// Order represents a customer order in the system.
type Order struct {
	ID     int
	Type   enum.OrderType
	Status enum.OrderStatus
}

// NewOrder creates a new order with the given ID and type, defaulting to Pending status.
func NewOrder(id int, orderType enum.OrderType) *Order {
	return &Order{
		ID:     id,
		Type:   orderType,
		Status: enum.Pending,
	}
}
