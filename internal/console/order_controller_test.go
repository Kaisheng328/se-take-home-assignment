package console

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"mcd-bot/internal/backend/enum"
	"mcd-bot/internal/backend/logger"
)

const testProcessingTime = 100 * time.Millisecond

func newTestController() (*OrderController, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	log := logger.New(buf)
	oc := NewOrderController(testProcessingTime, log)
	return oc, buf
}

func TestCreateOrderIncrementsID(t *testing.T) {
	oc, _ := newTestController()

	o1 := oc.NewOrder(enum.Normal)
	o2 := oc.NewOrder(enum.VIP)
	o3 := oc.NewOrder(enum.Normal)

	if o1.ID != 1 || o2.ID != 2 || o3.ID != 3 {
		t.Errorf("expected IDs 1,2,3 but got %d,%d,%d", o1.ID, o2.ID, o3.ID)
	}
}

func TestOrderTypeAssignment(t *testing.T) {
	oc, _ := newTestController()

	normal := oc.NewOrder(enum.Normal)
	vip := oc.NewOrder(enum.VIP)

	if normal.Type != enum.Normal {
		t.Errorf("expected Normal type, got %s", normal.Type)
	}
	if vip.Type != enum.VIP {
		t.Errorf("expected VIP type, got %s", vip.Type)
	}
}

func TestBotProcessesOrderOnCreation(t *testing.T) {
	oc, buf := newTestController()

	oc.NewOrder(enum.Normal)
	oc.AddBot()

	// Wait for processing to complete
	time.Sleep(200 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "picked up") {
		t.Errorf("expected bot to pick up order, output: %s", output)
	}
	if !strings.Contains(output, "completed") {
		t.Errorf("expected bot to complete order, output: %s", output)
	}

	completed := oc.GetCompletedOrders()
	if len(completed) != 1 {
		t.Errorf("expected 1 completed order, got %d", len(completed))
	}
}

func TestBotProcessesVIPFirst(t *testing.T) {
	oc, buf := newTestController()

	oc.NewOrder(enum.Normal)
	oc.NewOrder(enum.VIP)
	oc.AddBot()

	// Wait for first order to be picked up
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	// Bot should pick up VIP Order #2 first
	if !strings.Contains(output, "picked up VIP Order #2") {
		t.Errorf("expected bot to pick up VIP order first, output: %s", output)
	}
}

func TestRemoveBotReturnsOrderToQueue(t *testing.T) {
	oc, buf := newTestController()

	oc.NewOrder(enum.Normal)
	oc.AddBot()

	// Give bot time to start processing but not finish
	time.Sleep(30 * time.Millisecond)
	oc.RemoveBot()

	output := buf.String()
	if !strings.Contains(output, "returned to PENDING") {
		t.Errorf("expected order to return to pending, output: %s", output)
	}

	pending := oc.GetPendingOrders()
	if len(pending) != 1 {
		t.Errorf("expected 1 pending order after bot removal, got %d", len(pending))
	}
}

func TestRemoveIdleBot(t *testing.T) {
	oc, buf := newTestController()

	oc.AddBot()

	// Bot should be idle with no orders
	time.Sleep(20 * time.Millisecond)
	oc.RemoveBot()

	output := buf.String()
	if !strings.Contains(output, "destroyed while IDLE") {
		t.Errorf("expected idle bot destruction, output: %s", output)
	}
}

func TestRemoveBotFromEmptyList(t *testing.T) {
	oc, _ := newTestController()

	err := oc.RemoveBot()
	if err == nil {
		t.Error("expected error when removing bot from empty list")
	}
}

func TestIdleBotPicksUpNewOrder(t *testing.T) {
	oc, buf := newTestController()

	oc.AddBot()
	time.Sleep(20 * time.Millisecond)

	// Now add an order — idle bot should pick it up
	oc.NewOrder(enum.Normal)
	time.Sleep(200 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "picked up Normal Order #1") {
		t.Errorf("expected idle bot to pick up new order, output: %s", output)
	}

	completed := oc.GetCompletedOrders()
	if len(completed) != 1 {
		t.Errorf("expected 1 completed order, got %d", len(completed))
	}
}

func TestMultipleBotsProcessConcurrently(t *testing.T) {
	oc, _ := newTestController()

	oc.NewOrder(enum.Normal)
	oc.NewOrder(enum.VIP)
	oc.NewOrder(enum.Normal)

	oc.AddBot()
	oc.AddBot()

	// Wait enough for 2 orders to complete (but not all 3)
	time.Sleep(150 * time.Millisecond)

	completed := oc.GetCompletedOrders()
	if len(completed) < 2 {
		t.Errorf("expected at least 2 completed orders with 2 bots, got %d", len(completed))
	}

	// Wait for the last order
	time.Sleep(150 * time.Millisecond)

	completed = oc.GetCompletedOrders()
	if len(completed) != 3 {
		t.Errorf("expected 3 completed orders, got %d", len(completed))
	}
}

func TestLogContainsTimestamps(t *testing.T) {
	_, buf := newTestController()

	output := buf.String()
	// Logger should output timestamp in [HH:MM:SS] format
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Errorf("expected timestamps in output, got: %s", output)
	}
}

func TestGetStatusFormat(t *testing.T) {
	oc, _ := newTestController()

	oc.NewOrder(enum.Normal)
	oc.AddBot()
	time.Sleep(200 * time.Millisecond)

	status := oc.GetStatus()
	if !strings.Contains(status, "Final Status:") {
		t.Errorf("expected 'Final Status:' in status output, got: %s", status)
	}
	if !strings.Contains(status, "Orders Completed: 1") {
		t.Errorf("expected '1' completed order in status, got: %s", status)
	}
}
