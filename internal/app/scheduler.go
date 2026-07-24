package app

import (
	"context"
	"fmt"
	"sync"
)

type Work struct {
	Key      string
	Priority int
	Bytes    int64
	Run      func(context.Context) (any, error)
}
type WorkResult struct {
	Key   string
	Value any
	Err   error
}

// RunScheduled bounds concurrency and input bytes while returning deterministic key order.
func RunScheduled(ctx context.Context, workers int, byteBudget int64, work []Work) []WorkResult {
	if workers < 1 {
		workers = 1
	}
	if byteBudget < 1 {
		byteBudget = 1
	}
	items := append([]Work(nil), work...)
	sortWork(items)
	results := make([]WorkResult, len(items))
	tokens := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for i, item := range items {
		wg.Add(1)
		go func(i int, item Work) {
			defer wg.Done()
			select {
			case tokens <- struct{}{}:
			case <-ctx.Done():
				results[i] = WorkResult{Key: item.Key, Err: ctx.Err()}
				return
			}
			defer func() { <-tokens }()
			if item.Bytes > byteBudget {
				results[i] = WorkResult{Key: item.Key, Err: fmt.Errorf("work exceeds byte budget")}
				return
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						results[i] = WorkResult{Key: item.Key, Err: fmt.Errorf("worker panic recovered")}
					}
				}()
				value, err := item.Run(ctx)
				results[i] = WorkResult{Key: item.Key, Value: value, Err: err}
			}()
		}(i, item)
	}
	wg.Wait()
	return results
}

func sortWork(items []Work) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && (items[j].Priority > items[j-1].Priority || items[j].Priority == items[j-1].Priority && items[j].Key < items[j-1].Key); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}
