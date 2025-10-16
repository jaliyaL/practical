package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// v1 - Minimal Exponential Backoff Logic in Go
func BackoffRetry01() {
	rand.Seed(time.Now().UnixNano())

	baseDelay := 100 * time.Millisecond
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Simulate a random failure
		success := rand.Float64() > 0.5
		if success {
			fmt.Println("✅ Operation succeeded on attempt", attempt)
			break
		}

		// Exponential backoff delay
		delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
		fmt.Printf("❌ Attempt %d failed, retrying after %v\n", attempt, delay)
		time.Sleep(delay)
	}

	fmt.Println("Done retry loop")
}

// v2 - Modern Random Example for Backoff + jitter
func BackoffRetry02() {
	// Create a new random source and generator
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	baseDelay := 100 * time.Millisecond
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Simulate random failure
		success := r.Float64() > 0.5
		if success {
			fmt.Println("✅ Operation succeeded on attempt", attempt)
			break
		}

		// Exponential backoff delay
		delay := baseDelay * (1 << (attempt - 1)) // same as 2^(attempt-1)
		fmt.Printf("❌ Attempt %d failed, retrying after %v\n", attempt, delay)

		// Add jitter
		jitter := time.Duration(r.Int63n(int64(delay)))
		time.Sleep(jitter)
	}

	fmt.Println("Done retry loop")
}

// v3 - Minimal Concurrent Retry with Backoff
func BackoffRetry03() {
	rand.Seed(time.Now().UnixNano())

	numJobs := 5
	numWorkers := 3
	jobs := make(chan int, numJobs)
	var wg3 sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start workers
	for w := 1; w <= numWorkers; w++ {
		wg3.Add(1)
		go worker3(ctx, w, jobs, &wg3)
	}

	// Send jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	wg3.Wait()
	fmt.Println("✅ All jobs done or timed out")
}

// Worker function
func worker3(ctx context.Context, id int, jobs <-chan int, wg3 *sync.WaitGroup) {
	defer wg3.Done()

	for job := range jobs {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d stopping due to context\n", id)
			return
		default:
		}

		fmt.Printf("Worker %d starting job %d\n", id, job)
		if err := doJobWithRetry(ctx, job); err != nil {
			fmt.Printf("Worker %d: job %d failed after retries: %v\n", id, job, err)
		} else {
			fmt.Printf("Worker %d: job %d succeeded\n", id, job)
		}
	}
}

// Simulate a job with exponential backoff retry
func doJobWithRetry(ctx context.Context, job int) error {
	baseDelay := 100 * time.Millisecond
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("job %d canceled", job)
		default:
		}

		// Simulate random failure
		if rand.Float64() < 0.5 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			fmt.Printf("   Job %d failed, retry %d after %v\n", job, attempt, delay)
			time.Sleep(delay)
			continue
		}

		// Success
		return nil
	}

	return fmt.Errorf("job %d failed after %d retries", job, maxRetries)
}
