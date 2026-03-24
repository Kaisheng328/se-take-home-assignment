package enum

// OrderStatus represents the current status of an order.
type OrderStatus string

const (
	Pending    OrderStatus = "PENDING"
	Processing OrderStatus = "PROCESSING"
	Complete   OrderStatus = "COMPLETE"
)
