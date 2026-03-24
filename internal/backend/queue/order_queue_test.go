package queue

import (
	"testing"

	"mcd-bot/internal/backend/enum"
	"mcd-bot/internal/backend/model"
)

func TestEnqueueNormalOrders(t *testing.T) {
	q := NewOrderQueue()

	o1 := model.NewOrder(1, enum.Normal)
	o2 := model.NewOrder(2, enum.Normal)
	o3 := model.NewOrder(3, enum.Normal)

	q.Enqueue(o1)
	q.Enqueue(o2)
	q.Enqueue(o3)

	if q.Len() != 3 {
		t.Fatalf("expected queue length 3, got %d", q.Len())
	}

	// Should dequeue in FIFO order
	got := q.Dequeue()
	if got.ID != 1 {
		t.Errorf("expected order #1, got #%d", got.ID)
	}
	got = q.Dequeue()
	if got.ID != 2 {
		t.Errorf("expected order #2, got #%d", got.ID)
	}
	got = q.Dequeue()
	if got.ID != 3 {
		t.Errorf("expected order #3, got #%d", got.ID)
	}
}

func TestVIPOrdersPriority(t *testing.T) {
	q := NewOrderQueue()

	o1 := model.NewOrder(1, enum.Normal)
	o2 := model.NewOrder(2, enum.Normal)
	o3 := model.NewOrder(3, enum.VIP)

	q.Enqueue(o1)
	q.Enqueue(o2)
	q.Enqueue(o3)

	// VIP should be first
	got := q.Dequeue()
	if got.ID != 3 {
		t.Errorf("expected VIP order #3 first, got #%d", got.ID)
	}
	got = q.Dequeue()
	if got.ID != 1 {
		t.Errorf("expected normal order #1 second, got #%d", got.ID)
	}
	got = q.Dequeue()
	if got.ID != 2 {
		t.Errorf("expected normal order #2 third, got #%d", got.ID)
	}
}

func TestVIPOrdersQueueBehindExistingVIP(t *testing.T) {
	q := NewOrderQueue()

	o1 := model.NewOrder(1, enum.Normal)
	o2 := model.NewOrder(2, enum.VIP)
	o3 := model.NewOrder(3, enum.VIP)
	o4 := model.NewOrder(4, enum.Normal)

	q.Enqueue(o1)
	q.Enqueue(o2)
	q.Enqueue(o4)
	q.Enqueue(o3)

	// Expected order: VIP#2, VIP#3, Normal#1, Normal#4
	expected := []int{2, 3, 1, 4}
	for _, expectedID := range expected {
		got := q.Dequeue()
		if got.ID != expectedID {
			t.Errorf("expected order #%d, got #%d", expectedID, got.ID)
		}
	}
}

func TestDequeueFromEmptyQueue(t *testing.T) {
	q := NewOrderQueue()

	got := q.Dequeue()
	if got != nil {
		t.Errorf("expected nil from empty queue, got order #%d", got.ID)
	}
}

func TestInsertAtPriority(t *testing.T) {
	q := NewOrderQueue()

	o1 := model.NewOrder(1, enum.Normal)
	o2 := model.NewOrder(2, enum.VIP)
	o3 := model.NewOrder(3, enum.Normal)

	q.Enqueue(o1)
	q.Enqueue(o3)

	// Simulate a VIP order returning from a cancelled bot
	q.InsertAtPriority(o2)

	// VIP should still be first
	got := q.Dequeue()
	if got.ID != 2 {
		t.Errorf("expected VIP order #2 first, got #%d", got.ID)
	}
}

func TestIsEmpty(t *testing.T) {
	q := NewOrderQueue()

	if !q.IsEmpty() {
		t.Error("expected empty queue")
	}

	q.Enqueue(model.NewOrder(1, enum.Normal))
	if q.IsEmpty() {
		t.Error("expected non-empty queue")
	}

	q.Dequeue()
	if !q.IsEmpty() {
		t.Error("expected empty queue after dequeue")
	}
}
