package main

import (
	"fmt"
	"os"
	"time"

	"mcd-bot/internal/backend/enum"
	"mcd-bot/internal/backend/logger"
	"mcd-bot/internal/console"
)

func main() {
	log := logger.New(os.Stdout)

	fmt.Println("McDonald's Order Management System - Simulation Results")
	fmt.Println()

	processingTime := 10 * time.Second
	oc := console.NewOrderController(processingTime, log)

	// Step 1: Create initial orders
	oc.NewOrder(enum.Normal)
	time.Sleep(500 * time.Millisecond)

	oc.NewOrder(enum.VIP)
	time.Sleep(500 * time.Millisecond)

	oc.NewOrder(enum.Normal)
	time.Sleep(500 * time.Millisecond)

	// Step 2: Add bots — they should pick up VIP order first
	oc.AddBot()
	time.Sleep(500 * time.Millisecond)

	oc.AddBot()
	time.Sleep(500 * time.Millisecond)

	// Wait for bots to complete their first orders
	time.Sleep(10 * time.Second)

	// Step 3: Create another VIP order — idle bot should pick it up
	oc.NewOrder(enum.VIP)
	time.Sleep(500 * time.Millisecond)

	// Wait for remaining orders to complete
	time.Sleep(11 * time.Second)

	// Step 4: Remove a bot while idle
	oc.RemoveBot()
	time.Sleep(500 * time.Millisecond)

	// Print final status
	fmt.Println()
	fmt.Println(oc.GetStatus())
}
