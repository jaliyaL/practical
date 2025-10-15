package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task 01 - Run a goroutine with context.WithCancel and stop it after 2 seconds
func ContextTask01() {
	fmt.Println("task 01")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("worker stopped :", ctx.Err())
				return
			default:
				fmt.Println("working")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	time.Sleep(2 * time.Second)
	fmt.Println("cancel called")
	cancel()
	time.Sleep(1 * time.Second)
}

// Task 02 - Make a mock HTTP request that auto-cancels after timeout
func ContextTask02() {
	fmt.Println("task 02 started")

	// Create a context with 1-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Simulated worker that repeatedly "calls API"
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("worker stopped:", ctx.Err())
				return
			default:
				fmt.Println("calling API...")
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// Wait long enough to see cancellation in action
	time.Sleep(3 * time.Second)
	fmt.Println("main done")
}

// 🧩 Task 03 — Chain contexts between 2 functions (child cancels when parent does)
func ContextTask03() {
	fmt.Println("\n--- Task 03 started ---")

	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	go childWorker(parentCtx)

	time.Sleep(2 * time.Second)
	fmt.Println("calling parent cancel")
	parentCancel() // 🔴 triggers child cancel

	time.Sleep(1 * time.Second)
	fmt.Println("main done (task 03)")
}

// childWorker inherits cancellation from parent context
func childWorker(parent context.Context) {
	childCtx, cancel := context.WithCancel(parent)
	defer cancel()

	for {
		select {
		case <-childCtx.Done():
			fmt.Println("child stopped due to:", childCtx.Err())
			return
		default:
			fmt.Println("child working...")
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// Task 04 - worker pool where one failure cancels all others
func worker(ctx context.Context, w int, jobCh <-chan int, resultsCh chan<- int, cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d canceled: %v\n", w, ctx.Err())
			return

		case j, ok := <-jobCh:
			if !ok {
				// no more jobs
				return
			}

			// simulate work
			time.Sleep(500 * time.Millisecond)

			// simulate error on job 5
			if j == 5 {
				fmt.Printf("worker %d failed on job %d — canceling all workers\n", w, j)
				cancel() // 🔴 cancel all
				return
			}

			fmt.Printf("worker %d finished job %d\n", w, j)
			resultsCh <- j
		}
	}
}

func ContextTask04() {
	fmt.Println("=== Task 04: Worker pool with cancel on failure ===")

	numWorkers := 4
	numJobs := 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobCh := make(chan int)
	resultsCh := make(chan int)

	var wg sync.WaitGroup

	// start workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(ctx, w, jobCh, resultsCh, cancel, &wg)
	}

	// feed jobs
	go func() {
		defer close(jobCh)
		for j := 1; j <= numJobs; j++ {
			select {
			case <-ctx.Done():
				fmt.Println("🔴 stopping job dispatch due to cancel")
				return
			case jobCh <- j:
				time.Sleep(100 * time.Millisecond)

			}

		}

	}()

	// close results after workers finish
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// collect results
	for r := range resultsCh {
		fmt.Printf("✅ Job %d completed successfully\n", r)
	}

	fmt.Println("All done")
}

// Task 05 - Pipeline with context cancellation in any stage

// Stage 1: Generator with multiple workers
func generator(ctx context.Context, nums []int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case <-ctx.Done():
				fmt.Println("🚫 generator canceled")
				return
			case out <- n:
				fmt.Println("📤 generator produced:", n)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return out
}

// Stage 2: Processor with N workers
func processor(ctx context.Context, in <-chan int, numWorkers int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	worker := func(id int) {
		defer wg.Done()
		for n := range in {
			select {
			case <-ctx.Done():
				fmt.Printf("🚫 processor-%d canceled\n", id)
				return
			default:
				result := n * n
				fmt.Printf("🧮 processor-%d processed %d → %d\n", id, n, result)
				out <- result
				time.Sleep(200 * time.Millisecond)
			}
		}
	}

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Stage 3: Sink with multiple consumers
func sink(ctx context.Context, in <-chan int, cancel context.CancelFunc, numConsumers int) {
	var wg sync.WaitGroup

	consumer := func(id int) {
		defer wg.Done()
		for r := range in {
			select {
			case <-ctx.Done():
				fmt.Printf("🚫 sink-%d canceled\n", id)
				return
			default:
				if r == 16 {
					fmt.Printf("💥 sink-%d error on value %d — canceling pipeline\n", id, r)
					cancel()
					return
				}
				fmt.Printf("✅ sink-%d received: %d\n", id, r)
				time.Sleep(150 * time.Millisecond)
			}
		}
	}

	for i := 1; i <= numConsumers; i++ {
		wg.Add(1)
		go consumer(i)
	}
	//
	wg.Wait()
}

func ContextTask05() {
	fmt.Println("=== Multi-Worker Pipeline Example ===")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := []int{1, 2, 3, 4, 5, 6}

	gen := generator(ctx, nums)
	proc := processor(ctx, gen, 3) // 3 processor workers
	sink(ctx, proc, cancel, 2)     // 2 sink consumers

	time.Sleep(500 * time.Millisecond)
	fmt.Println("🏁 main done")
}
