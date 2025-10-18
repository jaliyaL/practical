package sliceandmap

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

func SM04() {
	rand.Seed(time.Now().UnixNano())

	numTasks := 20
	numWorkers := 5
	var results []int
	var mu sync.Mutex

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tasks := make(chan int, numTasks)
	var wg sync.WaitGroup

	// Worker function
	worker := func(id int, tasks <-chan int) {
		for n := range tasks {
			select {
			case <-ctx.Done():
				fmt.Printf("Worker %d canceled\n", id)
				return
			default:
				// Optional: simulate random failure
				if rand.Float32() < 0.1 {
					fmt.Printf("Worker %d task %d failed, retrying...\n", id, n)
					backoff(n)
				}

				// Compute result (square of the number)
				result := n * n
				time.Sleep(time.Duration(rand.Intn(200)+50) * time.Millisecond) // simulate work

				// Append to shared slice safely
				mu.Lock()
				results = append(results, result)
				mu.Unlock()

				fmt.Printf("Worker %d finished task %d -> result %d\n", id, n, result)
			}
		}
	}

	// Start workers
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, tasks)
		}(i)
	}

	// Send tasks
	for i := 1; i <= numTasks; i++ {
		select {
		case <-ctx.Done():
			break
		case tasks <- i:
		}
	}
	close(tasks)

	// Wait for all workers
	wg.Wait()

	// Sort results and print top 5
	sort.Sort(sort.Reverse(sort.IntSlice(results)))
	fmt.Println("\nTop 5 results:")
	for i := 0; i < 5 && i < len(results); i++ {
		fmt.Printf("%d ", results[i])
	}
	fmt.Println()
}

// Optional exponential backoff simulation
func backoff(task int) {
	delay := 100 * time.Millisecond
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		time.Sleep(delay)
		if rand.Float32() > 0.2 { // simulate success
			fmt.Printf("Task %d succeeded after retry %d\n", task, i+1)
			return
		}
		delay *= 2
	}
	fmt.Printf("Task %d failed after %d retries\n", task, maxRetries)
}
