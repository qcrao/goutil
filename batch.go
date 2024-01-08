package goutil

import (
	"context"
	"errors"
	"sync"
)

var ErrInvalidBatchType = errors.New("invalid batch type")

// BatchResult defines the interface for merging batch results.
type BatchResult interface {
	Merge(BatchResult)
}

// BatchProcessor defines the interface for batch processing.
type BatchProcessor interface {
	SplitBatch(items interface{}, batchSize int) []interface{}
	Process(ctx context.Context, batch interface{}) (BatchResult, error) // batch 本身是一个 slice
}

// BatchProcess performs the batch processing.
func BatchProcess(ctx context.Context, items interface{}, batchSize int, concurrency int, processor BatchProcessor) (BatchResult, error) {
	var wg sync.WaitGroup
	batches := processor.SplitBatch(items, batchSize)
	ch := make(chan BatchResult, len(batches))
	errCh := make(chan error, len(batches))
	sem := make(chan struct{}, concurrency)

	for _, batch := range batches {
		wg.Add(1)

		batch := batch
		Go(func() {
			defer wg.Done()
			defer func() { <-sem }() // Release a token

			sem <- struct{}{} // Acquire a token
			result, err := processor.Process(ctx, batch)
			if err != nil {
				errCh <- err
				return
			}
			ch <- result
		})
	}

	Go(func() {
		wg.Wait()
		close(ch)
		close(errCh)
	})

	var result BatchResult
	for r := range ch {
		if result == nil {
			result = r
		} else {
			result.Merge(r)
		}
	}

	if len(errCh) > 0 {
		return nil, <-errCh // Return the first error
	}

	return result, nil
}

// BatchSplitter defines the interface for splitting items into batches.
type BatchSplitter interface {
	SplitBatch(items interface{}, batchSize int) []interface{}
}

// Int64Split is used to split batches of type []int64.
type Int64Split struct{}

// SplitBatch splits the []int64 into batches.
func (s *Int64Split) SplitBatch(items interface{}, batchSize int) []interface{} {
	ids, ok := items.([]int64)
	if !ok {
		return nil
	}
	var batches []interface{}
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batches = append(batches, ids[i:end])
	}
	return batches
}
