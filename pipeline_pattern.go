package main

import (
	"context"
	"fmt"
	"time"
)

/*---------------------------------------
Let’s build a 3-stage pipeline:
Generator – produces numbers.
Square – squares each number.
Printer (sink) – prints results.
Each stage runs concurrently and passes data through channels.
-----------------------------------------*/

func generator1(ctx context.Context, nums ...int) <-chan int {

	out := make(chan int)

	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case <-ctx.Done():
				fmt.Println("generator canceled", ctx.Err())
				return
			case out <- n:
				fmt.Println("generator send ", n)
				time.Sleep(300 * time.Millisecond)

				// Wrap sleep in a cancel-aware select
				select {
				case <-ctx.Done():
					fmt.Println("generator sleep canceled:", ctx.Err())
					return
				case <-time.After(200 * time.Millisecond):
					// continue to next loop iteration
				}

			}
		}
	}()

	return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("square canceled", ctx.Err())
				return
			case r, ok := <-in:
				if !ok {

					return
				}
				out <- r * r
				fmt.Println("square send", r*r)
			}
		}
	}()

	return out
}

func printer(in <-chan int) {

	total := 0
	for i := range in {
		total += i
		fmt.Println("Printer received", i)
	}
	fmt.Println("Printer received total ", total)

}

func P01() {

	// Step 1: Create a base context with timeout
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelTimeout() // always good practice

	// Step 2: Wrap with manual cancel
	ctx, cancel := context.WithCancel(ctxTimeout)
	defer cancel()

	nums := generator1(ctx, 1, 2, 3, 4)
	squares := square(ctx, nums)
	go printer(squares)

	time.Sleep(500 * time.Millisecond)
	fmt.Println("cancelling jobs")
	cancel()

	time.Sleep(2 * time.Second)

}
