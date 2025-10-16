package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

/*---------------------------------
// -------  RATE LIMITER ------- //
------------------------------------
A Rate Limiter ensures that certain actions
(like API calls or requests) do not exceed a
 fixed rate — e.g. 5 requests per second.
--------------------------------*/
// v1 - basic version

func RateLimiterV1() {

	rate := 2 // allow 2 events per second
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	for i := 1; i <= 10; i++ {
		<-ticker.C
		println("Request", i, "at", time.Now().Format("15:04:05.000"))
	}
}

// v2 - Token Bucket Rate Limiter pattern
func RateLimiterV2() {

	rate := 3  // refill 3 tokens per second
	burst := 5 // allow up to 5 request at once
	tokens := make(chan struct{}, burst)

	// fill the bucket initially
	for i := 0; i < burst; i++ {
		tokens <- struct{}{}
	}

	// goroutine refills tokens at a steady rate
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rate))
		defer ticker.Stop()

		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			//added a token
			default:
				//bucket full - skills
			}
		}
	}()

	// simulate 10 requests
	for i := 1; i <= 10; i++ {
		<-tokens // wait for token
		fmt.Println("Request", i, "at", time.Now().Format("15:04:05.000"))
		time.Sleep(300 * time.Millisecond) // simulate small delay between requests
	}
}

// v3 - Example: Token Bucket Rate Limiter with Context - timeout

func RateLimiterV3() {
	rate := 2  // refill 2 tokens per second
	burst := 5 // allow up to 5 requests instantly
	tokens := make(chan struct{}, burst)

	// Fill the bucket initially
	for i := 0; i < burst; i++ {
		tokens <- struct{}{}
	}

	// Create a context that auto-cancels after 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Goroutine to refill tokens
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rate))
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Rate limiter stopped:", ctx.Err())
				return
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
					// added a token
				default:
					// bucket full — skip
				}
			}
		}
	}()

	// Simulate requests
	for i := 1; i <= 10; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Request loop stopped:", ctx.Err())
			return
		case <-tokens:
			fmt.Println("Request", i, "at", time.Now().Format("15:04:05.000"))
			time.Sleep(300 * time.Millisecond)
		}
	}
}

//v4 - Token Bucket Rate Limiter with Manual Cancel

func RateLimiterV4() {
	rate := 2  // refill 2 tokens per second
	burst := 5 // allow up to 5 requests at once
	tokens := make(chan struct{}, burst)

	// fill the bucket initially
	for i := 0; i < burst; i++ {
		tokens <- struct{}{}
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure cleanup

	// Refill goroutine
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rate))
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Refiller stopped:", ctx.Err())
				return
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
					// token added
				default:
					// bucket full
				}
			}
		}
	}()

	// Simulate requests
	for i := 1; i <= 10; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Request loop stopped:", ctx.Err())
			return
		case <-tokens:
			fmt.Println("Request", i, "at", time.Now().Format("15:04:05.000"))
			time.Sleep(300 * time.Millisecond)

			// Manually cancel after 6 requests
			if i == 6 {
				fmt.Println("⚠️  Canceling context manually after 6 requests")
				cancel()
			}
		}
	}
}

// v5 - Token Bucket + Context + Exponential Backoff
func RateLimiterV5() {
	rate := 2  // refill 2 tokens per second
	burst := 5 // allow 5 at once
	tokens := make(chan struct{}, burst)

	// Fill bucket initially
	for i := 0; i < burst; i++ {
		tokens <- struct{}{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	// Refill tokens at constant rate
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rate))
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Refiller stopped:", ctx.Err())
				return
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default:
					// bucket full
				}
			}
		}
	}()

	// Simulate 10 requests, some might fail and retry with backoff
	for i := 1; i <= 10; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Request loop stopped:", ctx.Err())
			return
		case <-tokens:
			fmt.Println("Request", i, "at", time.Now().Format("15:04:05.000"))

			// simulate possible failure for even-numbered requests
			if i%2 == 0 {
				fmt.Println("❌ Request", i, "failed — applying backoff...")
				backoffRetry(ctx, i)
			}

			time.Sleep(300 * time.Millisecond)
		}
	}
}

// backoffRetry simulates retrying a failed request with exponential backoff
func backoffRetry(ctx context.Context, id int) {
	baseDelay := 100 * time.Millisecond
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			fmt.Println("Backoff canceled for request", id)
			return
		default:
		}

		delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
		fmt.Printf("   ↪ Retry %d for Request %d after %v\n", attempt, id, delay)
		time.Sleep(delay)

		// simulate success after 3 tries
		if attempt == 3 {
			fmt.Printf("   ✅ Request %d succeeded after %d retries\n", id, attempt)
			return
		}
	}
	fmt.Printf("   ❌ Request %d failed after all retries\n", id)
}

// v6 - Concurrent Requests + Rate Limiter + Exponential Backoff

func RateLimiterV6() {
	rate := 2         // refill 2 tokens per second
	burst := 5        // allow short bursts up to 5
	numRequests := 10 // total simulated requests
	numWorkers := 3   // concurrent workers

	tokens := make(chan struct{}, burst)
	for i := 0; i < burst; i++ {
		tokens <- struct{}{} // prefill bucket
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Refill goroutine
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rate))
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Refiller stopped:", ctx.Err())
				return
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default: // bucket full
				}
			}
		}
	}()

	// Job channel
	jobs := make(chan int)

	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker6(ctx, w, tokens, jobs, &wg)
	}

	// Send jobs
	go func() {
		for i := 1; i <= numRequests; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- i:
			}
		}
		close(jobs)
	}()

	wg.Wait()
	fmt.Println("✅ All workers done or canceled.")
}

// Worker consumes tokens and performs "requests"
func worker6(ctx context.Context, id int, tokens <-chan struct{}, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-ctx.Done():
			fmt.Println("Worker", id, "stopping:", ctx.Err())
			return
		case <-tokens:
			fmt.Printf("Worker %d handling request %d at %s\n",
				id, job, time.Now().Format("15:04:05.000"))

			// Simulate random failure
			if rand.Float64() < 0.4 {
				fmt.Printf("❌ Worker %d: request %d failed — retrying with backoff\n", id, job)
				backoffRetry6(ctx, id, job)
			}

			time.Sleep(300 * time.Millisecond)
		}
	}
}

// Exponential backoff with jitter and max delay
func backoffRetry6(ctx context.Context, workerID, jobID int) {
	baseDelay := 100 * time.Millisecond
	maxDelay := 2 * time.Second
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			fmt.Printf("   ↪ Worker %d: retry for job %d canceled\n", workerID, jobID)
			return
		default:
		}

		// Exponential backoff delay
		delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
		if delay > maxDelay {
			delay = maxDelay
		}

		// Add jitter (randomization)
		jitter := time.Duration(rand.Int63n(int64(delay)))
		fmt.Printf("   ↪ Worker %d: retry %d for job %d after %v\n", workerID, attempt, jobID, jitter)
		time.Sleep(jitter)

		// Simulate success on 3rd attempt
		if attempt == 3 {
			fmt.Printf("   ✅ Worker %d: job %d succeeded after %d retries\n", workerID, jobID, attempt)
			return
		}
	}

	fmt.Printf("   ❌ Worker %d: job %d failed after all retries\n", workerID, jobID)
}
