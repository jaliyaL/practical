package sliceandmap

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

func SM05() {
	numIntegers := 100
	batchSize := 25 // 100 / 4
	numWorkers := 4

	// Generate dataset: 1..100
	data := make([]int, numIntegers)
	for i := 0; i < numIntegers; i++ {
		data[i] = i + 1
	}

	// Channel to send batches to workers
	batchCh := make(chan []int)
	// Channel to collect worker results
	mapCh := make(chan map[int]int, numWorkers)

	var wg sync.WaitGroup

	// Optional rate limiter: 1 batch per 200ms
	rateLimiter := time.Tick(200 * time.Millisecond)

	// Stage 1 + 2: Fan-out to workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for batch := range batchCh {
				<-rateLimiter // rate limiting

				// Compute map of counts for this batch
				localMap := make(map[int]int)
				for _, n := range batch {
					localMap[n] = n // for demo, value = number itself
				}

				fmt.Printf("Worker %d processed batch: %v\n", id, batch)
				mapCh <- localMap
			}
		}(i)
	}

	// Split data into batches and send to batchCh
	go func() {
		for i := 0; i < numIntegers; i += batchSize {
			end := i + batchSize
			if end > numIntegers {
				end = numIntegers
			}
			batchCh <- data[i:end]
		}
		close(batchCh)
	}()

	// Wait for workers in separate goroutine
	go func() {
		wg.Wait()
		close(mapCh)
	}()

	// Stage 3: Merge all maps
	finalMap := make(map[int]int)
	for m := range mapCh {
		for k, v := range m {
			finalMap[k] = v
		}
	}

	// Stage 4: Collect keys > 50 into slice and sort
	var slice []int
	for k := range finalMap {
		if k > 50 {
			slice = append(slice, k)
		}
	}
	sort.Ints(slice)

	// Print results
	fmt.Println("\nFinal aggregated map size:", len(finalMap))
	fmt.Println("Keys > 50 sorted:", slice)
}
