package queue

import (
	"mcd-bot/internal/backend/enum"
	"mcd-bot/internal/backend/model"
)

// OrderQueue is a priority queue that places VIP orders before Normal orders,
// while maintaining FIFO ordering within the same type.
type OrderQueue struct {
	orders []*model.Order
}

// NewOrderQueue creates an empty order queue.
func NewOrderQueue() *OrderQueue {
	return &OrderQueue{
		orders: make([]*model.Order, 0),
	}
}

// Enqueue adds an order to the queue with VIP priority.
// VIP orders are placed after existing VIP orders but before all Normal orders.
func (q *OrderQueue) Enqueue(order *model.Order) {
	if order.Type == enum.VIP {
		insertIdx := q.findLastVIPIndex() + 1
		q.insertAt(insertIdx, order)
	} else {
		q.orders = append(q.orders, order)
	}
}

// Dequeue removes and returns the first order from the queue.
// Returns nil if the queue is empty.
func (q *OrderQueue) Dequeue() *model.Order {
	if len(q.orders) == 0 {
		return nil
	}

	order := q.orders[0]
	q.orders = q.orders[1:]
	return order
}

// InsertAtPriority re-inserts a returned order at its correct priority position.
// Used when a bot is removed mid-processing and the order must return to the queue.
func (q *OrderQueue) InsertAtPriority(order *model.Order) {
	q.Enqueue(order)
}

// Len returns the number of orders in the queue.
func (q *OrderQueue) Len() int {
	return len(q.orders)
}

// GetAll returns a snapshot of all orders in the queue.
func (q *OrderQueue) GetAll() []*model.Order {
	result := make([]*model.Order, len(q.orders))
	copy(result, q.orders)
	return result
}

// IsEmpty returns true if the queue has no orders.
func (q *OrderQueue) IsEmpty() bool {
	return len(q.orders) == 0
}

// findLastVIPIndex returns the index of the last VIP order in the queue.
// Returns -1 if there are no VIP orders.
func (q *OrderQueue) findLastVIPIndex() int {
	lastVIP := -1
	for i, order := range q.orders {
		if order.Type == enum.VIP {
			lastVIP = i
		}
	}
	return lastVIP
}

// insertAt inserts an order at the specified index.
func (q *OrderQueue) insertAt(idx int, order *model.Order) {
	q.orders = append(q.orders, nil)
	copy(q.orders[idx+1:], q.orders[idx:])
	q.orders[idx] = order
}
