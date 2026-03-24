package console

import (
	"fmt"
	"sync"
	"time"

	"mcd-bot/internal/backend/enum"
	"mcd-bot/internal/backend/logger"
	"mcd-bot/internal/backend/model"
	"mcd-bot/internal/backend/queue"
)

// OrderController orchestrates order management and bot lifecycle.
type OrderController struct {
	mu              sync.Mutex
	pendingQueue    *queue.OrderQueue
	completedOrders []*model.Order
	bots            []*model.Bot
	nextOrderID     int
	nextBotID       int
	processingTime  time.Duration
	logger          *logger.Logger
}

// NewOrderController creates a new controller with the specified processing time and logger.
func NewOrderController(processingTime time.Duration, log *logger.Logger) *OrderController {
	oc := &OrderController{
		pendingQueue:    queue.NewOrderQueue(),
		completedOrders: make([]*model.Order, 0),
		bots:            make([]*model.Bot, 0),
		nextOrderID:     1,
		nextBotID:       1,
		processingTime:  processingTime,
		logger:          log,
	}

	log.Log("System initialized with 0 bots")
	return oc
}

// NewOrder creates a new order of the given type, adds it to the pending queue,
// and assigns it to an idle bot if one is available.
func (oc *OrderController) NewOrder(orderType enum.OrderType) *model.Order {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	order := model.NewOrder(oc.nextOrderID, orderType)
	oc.nextOrderID++

	oc.pendingQueue.Enqueue(order)
	oc.logger.Log("Created %s Order #%d - Status: %s", orderType, order.ID, order.Status)

	oc.assignIdleBots()
	return order
}

// AddBot creates a new bot and immediately assigns it a pending order if available.
func (oc *OrderController) AddBot() *model.Bot {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	bot := model.NewBot(oc.nextBotID)
	oc.nextBotID++
	oc.bots = append(oc.bots, bot)

	oc.logger.Log("Bot #%d created - Status: %s", bot.ID, enum.Idle)
	oc.tryAssignOrder(bot)
	return bot
}

// RemoveBot removes the newest bot. If it's processing an order,
// the order is cancelled and returned to the pending queue at its correct priority position.
func (oc *OrderController) RemoveBot() error {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if len(oc.bots) == 0 {
		oc.logger.Log("ERROR: No bots to remove")
		return fmt.Errorf("no bots to remove")
	}

	// Remove the newest bot (last in slice)
	bot := oc.bots[len(oc.bots)-1]
	oc.bots = oc.bots[:len(oc.bots)-1]

	// Cancel processing and return order to queue if applicable
	returnedOrder := bot.Cancel()
	if returnedOrder != nil {
		oc.pendingQueue.InsertAtPriority(returnedOrder)
		oc.logger.Log("Bot #%d destroyed while processing %s Order #%d - Order returned to PENDING",
			bot.ID, returnedOrder.Type, returnedOrder.ID)
	} else {
		oc.logger.Log("Bot #%d destroyed while %s", bot.ID, enum.Idle)
	}

	return nil
}

// GetPendingOrders returns a snapshot of the current pending orders.
func (oc *OrderController) GetPendingOrders() []*model.Order {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	return oc.pendingQueue.GetAll()
}

// GetCompletedOrders returns a snapshot of all completed orders.
func (oc *OrderController) GetCompletedOrders() []*model.Order {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	result := make([]*model.Order, len(oc.completedOrders))
	copy(result, oc.completedOrders)
	return result
}

// GetBots returns a snapshot of all active bots.
func (oc *OrderController) GetBots() []*model.Bot {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	result := make([]*model.Bot, len(oc.bots))
	copy(result, oc.bots)
	return result
}

// GetStatus returns a formatted string of the current system state.
func (oc *OrderController) GetStatus() string {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	vipCompleted := 0
	normalCompleted := 0
	for _, o := range oc.completedOrders {
		if o.Type == enum.VIP {
			vipCompleted++
		} else {
			normalCompleted++
		}
	}

	return fmt.Sprintf(
		"Final Status:\n"+
			"  - Total Orders Processed: %d (%d VIP, %d Normal)\n"+
			"  - Orders Completed: %d\n"+
			"  - Active Bots: %d\n"+
			"  - Pending Orders: %d",
		len(oc.completedOrders), vipCompleted, normalCompleted,
		len(oc.completedOrders),
		len(oc.bots),
		oc.pendingQueue.Len(),
	)
}

// assignIdleBots iterates through all bots and assigns pending orders to idle ones.
// Must be called with mu locked.
func (oc *OrderController) assignIdleBots() {
	for _, bot := range oc.bots {
		if bot.Status == enum.Idle && !oc.pendingQueue.IsEmpty() {
			oc.tryAssignOrder(bot)
		}
	}
}

// tryAssignOrder attempts to assign the next pending order to the given bot.
// Must be called with mu locked.
func (oc *OrderController) tryAssignOrder(bot *model.Bot) {
	order := oc.pendingQueue.Dequeue()
	if order == nil {
		oc.logger.Log("Bot #%d is now %s - No pending orders", bot.ID, enum.Idle)
		return
	}

	oc.logger.Log("Bot #%d picked up %s Order #%d - Status: %s", bot.ID, order.Type, order.ID, enum.Processing)
	bot.Process(order, oc.processingTime, oc.onOrderComplete)
}

// onOrderComplete is the callback invoked when a bot finishes processing an order.
func (oc *OrderController) onOrderComplete(bot *model.Bot, order *model.Order) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.completedOrders = append(oc.completedOrders, order)
	oc.logger.Log("Bot #%d completed %s Order #%d - Status: %s (Processing time: %ds)",
		bot.ID, order.Type, order.ID, order.Status, int(oc.processingTime.Seconds()))

	// Try to pick up the next pending order
	oc.tryAssignOrder(bot)
}
