package cache

import (
	"sync"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/velvetriddles/mortgage-calc/internal/model"
)

func TestMortCache_Save_GetAll(t *testing.T) {
	// Create a cache instance
	cache := NewMortCache()

	// Create test data
	resp := model.ExecuteResponse{
		Params: model.RequestParams{
			ObjectCost:     decimal.NewFromInt(5000000),
			InitialPayment: decimal.NewFromInt(1000000),
			Months:         240,
		},
		Program: model.ProgramRequest{
			Salary: true,
		},
		Aggregates: model.Aggregates{
			Rate:            decimal.NewFromFloat(8.0),
			LoanSum:         decimal.NewFromInt(4000000),
			MonthlyPayment:  decimal.NewFromInt(33458),
			Overpayment:     decimal.NewFromInt(4029920),
			LastPaymentDate: "2044-02-18",
		},
	}

	// Save data to cache
	id := cache.Save(resp)

	// Check that ID starts with 0
	if id != 0 {
		t.Errorf("Expected ID 0, got %d", id)
	}

	// Get all items from cache
	items, err := cache.GetAll()
	if err != nil {
		t.Fatalf("Unexpected error when getting data from cache: %v", err)
	}

	// Check number of items
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	// Check ID and item data
	if items[0].ID != 0 {
		t.Errorf("Expected ID 0, got %d", items[0].ID)
	}
}

func TestMortCache_Clear_Size(t *testing.T) {
	// Create a cache instance
	cache := NewMortCache()

	// Create test data
	resp := model.ExecuteResponse{
		Params: model.RequestParams{
			ObjectCost:     decimal.NewFromInt(5000000),
			InitialPayment: decimal.NewFromInt(1000000),
			Months:         240,
		},
		Program: model.ProgramRequest{
			Salary: true,
		},
		Aggregates: model.Aggregates{
			Rate:            decimal.NewFromFloat(8.0),
			LoanSum:         decimal.NewFromInt(4000000),
			MonthlyPayment:  decimal.NewFromInt(33458),
			Overpayment:     decimal.NewFromInt(4029920),
			LastPaymentDate: "2044-02-18",
		},
	}

	// Save data to cache
	cache.Save(resp)
	cache.Save(resp)

	// Check cache size
	if size := cache.Size(); size != 2 {
		t.Errorf("Expected cache size 2, got %d", size)
	}

	// Clear cache
	cache.Clear()

	// Check cache size after clearing
	if size := cache.Size(); size != 0 {
		t.Errorf("Expected cache size 0 after clearing, got %d", size)
	}

	// Check that GetAll returns empty cache error
	_, err := cache.GetAll()
	if err != ErrEmptyCache {
		t.Errorf("Expected error 'cache is empty', got: %v", err)
	}
}

func TestMortCache_Parallel(t *testing.T) {
	// Create a cache instance
	cache := NewMortCache()

	// Number of goroutines for parallel testing
	const goroutines = 10

	// Create test data
	resp := model.ExecuteResponse{
		Params: model.RequestParams{
			ObjectCost:     decimal.NewFromInt(5000000),
			InitialPayment: decimal.NewFromInt(1000000),
			Months:         240,
		},
		Program: model.ProgramRequest{
			Salary: true,
		},
		Aggregates: model.Aggregates{
			Rate:            decimal.NewFromFloat(8.0),
			LoanSum:         decimal.NewFromInt(4000000),
			MonthlyPayment:  decimal.NewFromInt(33458),
			Overpayment:     decimal.NewFromInt(4029920),
			LastPaymentDate: "2044-02-18",
		},
	}

	// Create WaitGroup for goroutine synchronization
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Launch goroutines for parallel cache saving
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cache.Save(resp)
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Get all items from cache
	items, err := cache.GetAll()
	if err != nil {
		t.Fatalf("Unexpected error when getting data from cache: %v", err)
	}

	// Check number of items
	if len(items) != goroutines {
		t.Errorf("Expected %d items, got %d", goroutines, len(items))
	}

	// Check that all IDs from 0 to goroutines-1 are present in the cache
	idMap := make(map[int]bool)
	for _, item := range items {
		idMap[item.ID] = true
	}

	for i := 0; i < goroutines; i++ {
		if !idMap[i] {
			t.Errorf("ID %d is missing from cache", i)
		}
	}
}
