package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/velvetriddles/mortgage-calc/internal/cache"
	"github.com/velvetriddles/mortgage-calc/internal/model"
	"github.com/velvetriddles/mortgage-calc/internal/service"
)

// TestExecuteHandler_Success tests a successful POST request to /execute
func TestExecuteHandler_Success(t *testing.T) {
	// Create test dependencies
	mortCache := cache.NewMortCache()
	calculator := service.NewMortCalculator()
	handler := NewMortHandler(mortCache, calculator)

	// Create test request
	reqBody := model.ExecuteRequest{
		ObjectCost:     decimal.NewFromInt(5000000),
		InitialPayment: decimal.NewFromInt(1000000),
		Months:         240,
		Program: model.ProgramRequest{
			Salary: true,
		},
	}

	// Serialize request to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Error marshaling request: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/execute", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Process the request
	handler.Execute(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Check response structure
	var resp SuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	// Check that monthly payment matches expected value
	expected := decimal.NewFromInt(33458)
	if !resp.Result.Aggregates.MonthlyPayment.Equal(expected) {
		t.Errorf("Expected monthly payment %v, got %v", expected, resp.Result.Aggregates.MonthlyPayment)
	}

	// Check that monthly payment is rounded to an integer
	rounded := resp.Result.Aggregates.MonthlyPayment.Round(0)
	if !resp.Result.Aggregates.MonthlyPayment.Equal(rounded) {
		t.Errorf("Monthly payment %v is not rounded to an integer", resp.Result.Aggregates.MonthlyPayment)
	}
}

// TestExecuteHandler_MultiplePrograms tests an error when multiple programs are selected
func TestExecuteHandler_MultiplePrograms(t *testing.T) {
	// Create test dependencies
	mortCache := cache.NewMortCache()
	calculator := service.NewMortCalculator()
	handler := NewMortHandler(mortCache, calculator)

	// Create test request with two programs
	reqBody := model.ExecuteRequest{
		ObjectCost:     decimal.NewFromInt(5000000),
		InitialPayment: decimal.NewFromInt(1000000),
		Months:         240,
		Program: model.ProgramRequest{
			Salary:   true,
			Military: true,
		},
	}

	// Serialize request to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Error marshaling request: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/execute", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Process the request
	handler.Execute(rr, req)

	// Check status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// Check error content
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	// Check error text according to requirements
	expectedError := "choose only 1 program"
	if resp.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, resp.Error)
	}
}

// TestGetCacheHandler_EmptyCache tests the error when requesting an empty cache
func TestGetCacheHandler_EmptyCache(t *testing.T) {
	// Create test dependencies with an empty cache
	mortCache := cache.NewMortCache()

	// Create a mock calculator implementing the Calculator interface
	mockCalculator := &MockCalculator{}

	handler := NewMortHandler(mortCache, mockCalculator)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/cache", nil)

	// Create ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Process the request
	handler.GetCache(rr, req)

	// Check status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// Check error content
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	// Check error text according to requirements
	expectedError := "empty cache"
	if resp.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, resp.Error)
	}
}

// MockCalculator - mock implementation of the Calculator interface for testing
type MockCalculator struct{}

// Calculate implements the Calculator interface method
func (m *MockCalculator) Calculate(req model.ExecuteRequest, baseTime time.Time) (model.Aggregates, error) {
	// Just return an empty structure for the test
	return model.Aggregates{}, nil
}
