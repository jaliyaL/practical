package main

import (
	"fmt"
	"sync"
	"time"
)

/*----------------------------------------
If you implement bounded parallelism correctly:
You do NOT start 1000 goroutines at once.
You start at most 5 goroutines at a time (the “concurrency limit”).
Once one goroutine finishes, the next job can start.
------------------------------------------*/

func BoundedParallelism() {

	const totalTasks = 20
	const maxConcurrent = 5

	// Semaphore to limit concurrency
	sem := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup

	for i := 1; i <= totalTasks; i++ {
		wg.Add(1)
		sem <- struct{}{} // acquire a slot

		go func(task int) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			// Simulate work
			fmt.Println("Starting task", task)
			time.Sleep(100 * time.Millisecond) // simulate work
			fmt.Println("Finished task", task)
		}(i)
	}

	// Wait for all tasks to complete
	wg.Wait()
	fmt.Println("All tasks done")
}
