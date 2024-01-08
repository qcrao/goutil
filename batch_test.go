package goutil

import (
	"context"
	"errors"
	"testing"
)

// TestInt64SplitWithInvalidType tests the Int64Split.SplitBatch method with an invalid type.
func TestInt64SplitWithInvalidType(t *testing.T) {
	splitter := &Int64Split{}
	items := "not a slice of int64" // Intentionally using an incorrect type

	batches := splitter.SplitBatch(items, 2)
	if batches != nil {
		t.Errorf("Expected nil, got %v", batches)
	}
}

// SumResult is a simple implementation of BatchResult that holds a sum of int64 values.
type SumResult struct {
	Sum int64
}

// Merge combines the result of another SumResult into this one.
func (r *SumResult) Merge(other BatchResult) {
	o, ok := other.(*SumResult)
	if !ok {
		return
	}
	r.Sum += o.Sum
}

// SumProcessor is a mock BatchProcessor that processes slices of int64 and returns their sum.
type SumProcessor struct{}

func (p *SumProcessor) SplitBatch(items interface{}, batchSize int) []interface{} {
	// Use Int64Split to split the batch
	splitter := &Int64Split{}
	return splitter.SplitBatch(items, batchSize)
}

func (p *SumProcessor) Process(ctx context.Context, batch interface{}) (BatchResult, error) {
	nums, ok := batch.([]int64)
	if !ok {
		return nil, ErrInvalidBatchType
	}

	sum := int64(0)
	for _, num := range nums {
		sum += num
	}
	return &SumResult{Sum: sum}, nil
}

// TestBatchProcessWithInt64 tests the BatchProcess function with int64 slices.
func TestBatchProcessWithInt64(t *testing.T) {
	ctx := context.TODO()
	items := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	batchSize := 3
	concurrency := 2
	processor := &SumProcessor{}

	result, err := BatchProcess(ctx, items, batchSize, concurrency, processor)
	if err != nil {
		t.Errorf("BatchProcess returned an error: %v", err)
	}

	// Assert the result
	expectedSum := int64(55) // sum of numbers from 1 to 10
	sumResult, ok := result.(*SumResult)
	if !ok || sumResult.Sum != expectedSum {
		t.Errorf("Expected sum %v, got %v", expectedSum, sumResult.Sum)
	}
}

// ErrorSumProcessor is a mock BatchProcessor that intentionally returns an error for testing.
type ErrorSumProcessor struct {
	ErrorValue int64 // Value that triggers an error
}

func (p *ErrorSumProcessor) SplitBatch(items interface{}, batchSize int) []interface{} {
	splitter := &Int64Split{}
	return splitter.SplitBatch(items, batchSize)
}

func (p *ErrorSumProcessor) Process(ctx context.Context, batch interface{}) (BatchResult, error) {
	nums, ok := batch.([]int64)
	if !ok {
		return nil, ErrInvalidBatchType
	}

	for _, num := range nums {
		if num == p.ErrorValue {
			return nil, errors.New("error processing batch")
		}
	}
	return &SumResult{}, nil
}

// TestBatchProcessWithError tests BatchProcess when the processor returns an error.
func TestBatchProcessWithError(t *testing.T) {
	ctx := context.TODO()
	items := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	batchSize := 3
	concurrency := 2
	errorValue := int64(5) // Set the error trigger value
	processor := &ErrorSumProcessor{ErrorValue: errorValue}

	_, err := BatchProcess(ctx, items, batchSize, concurrency, processor)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}
