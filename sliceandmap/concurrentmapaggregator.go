package sliceandmap

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/*-------------------------
1. Concurrent Map Aggregator
Scenario:
You have multiple simulated APIs returning key-value results concurrently.
Task:
•	Spawn 10 goroutines to fetch data (simulate with sleep + random data).
•	Each goroutine writes its result to a shared map (map[string]int).
•	Protect the map with a sync.Mutex or use sync.Map.
•	Merge results safely and print the final aggregated map.
•	Cancel all goroutines if any goroutine fails using context.WithCancel.
---------------------------*/

func SM01() {

	// Shared map and mutex
	resultMap := make(map[string]int)
	var mu sync.Mutex

	// Success and failure counters
	var success, failure int64
	var counterMu sync.Mutex

	// Context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Simulate API call with sleep
			delay := time.Duration(100+id*50) * time.Millisecond // deterministic delay
			select {
			case <-ctx.Done():
				fmt.Printf("Goroutine %d canceled\n", id)
				return
			case <-time.After(delay):
				// Simulate failure for example (let's fail goroutine 5)
				if id == 5 {
					fmt.Printf("Goroutine %d failed\n", id)
					counterMu.Lock()
					failure++
					counterMu.Unlock()
					cancel() // cancel all goroutines
					return
				}

				// Use goroutine number as key & value
				key := fmt.Sprintf("goroutine-%d", id)
				value := id

				// Safely write to shared map
				mu.Lock()
				resultMap[key] = value
				mu.Unlock()

				counterMu.Lock()
				success++
				counterMu.Unlock()

				fmt.Printf("Goroutine %d succeeded: %s=%d\n", id, key, value)
			}
		}(i)
	}

	wg.Wait()

	fmt.Println("\n--- Final Aggregated Map ---")
	for k, v := range resultMap {
		fmt.Printf("%s: %d\n", k, v)
	}

	fmt.Printf("\nSuccess: %d, Failure: %d\n", success, failure)
}

// Thread-safe map

func SM02() {
	// Thread-safe map
	var resultMap sync.Map

	// Success and failure counters
	var success, failure int64

	// Context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Simulate API call with deterministic delay
			delay := time.Duration(100+id*50) * time.Millisecond
			select {
			case <-ctx.Done():
				fmt.Printf("Goroutine %d canceled\n", id)
				return
			case <-time.After(delay):
				// Simulate failure for goroutine 5
				if id == 5 {
					fmt.Printf("Goroutine %d failed\n", id)
					atomic.AddInt64(&failure, 1)
					cancel() // cancel all goroutines
					return
				}

				// Use goroutine number as key & value
				key := fmt.Sprintf("goroutine-%d", id)
				value := id

				// Store in sync.Map (thread-safe)
				resultMap.Store(key, value)

				// Increment success counter
				atomic.AddInt64(&success, 1)

				fmt.Printf("Goroutine %d succeeded: %s=%d\n", id, key, value)
			}
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()

	// Print aggregated map
	fmt.Println("\n--- Final Aggregated Map ---")
	resultMap.Range(func(k, v any) bool {
		fmt.Printf("%s: %v\n", k, v)
		return true
	})

	fmt.Printf("\nSuccess: %d, Failure: %d\n", success, failure)
}
