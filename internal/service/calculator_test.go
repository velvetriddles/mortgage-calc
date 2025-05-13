package service

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/velvetriddles/mortgage-calc/internal/model"
)

func TestCalculate_HappyPath(t *testing.T) {
	// Fix time for tests
	baseTime := time.Date(2024, 2, 18, 12, 0, 0, 0, time.UTC)
	calculator := NewMortCalculator()

	tests := []struct {
		name        string
		request     model.ExecuteRequest
		expected    model.Aggregates
		expectedErr error
	}{
		{
			name: "Corporate client program (salary)",
			request: model.ExecuteRequest{
				ObjectCost:     decimal.NewFromInt(5000000),
				InitialPayment: decimal.NewFromInt(1000000), // 20%
				Months:         240,
				Program: model.ProgramRequest{
					Salary: true,
				},
			},
			expected: model.Aggregates{
				Rate:            decimal.NewFromFloat(8.0),
				LoanSum:         decimal.NewFromInt(4000000),
				MonthlyPayment:  decimal.NewFromInt(33458),
				Overpayment:     decimal.NewFromInt(4029920),
				LastPaymentDate: "2044-02-18",
			},
			expectedErr: nil,
		},
		{
			name: "Military mortgage (military)",
			request: model.ExecuteRequest{
				ObjectCost:     decimal.NewFromInt(8000000),
				InitialPayment: decimal.NewFromInt(2000000), // 25%
				Months:         180,
				Program: model.ProgramRequest{
					Military: true,
				},
			},
			expected: model.Aggregates{
				Rate:            decimal.NewFromFloat(9.0),
				LoanSum:         decimal.NewFromInt(6000000),
				MonthlyPayment:  decimal.NewFromInt(60856),
				Overpayment:     decimal.NewFromInt(4954080),
				LastPaymentDate: "2039-02-18",
			},
			expectedErr: nil,
		},
		{
			name: "Base program (base)",
			request: model.ExecuteRequest{
				ObjectCost:     decimal.NewFromInt(3000000),
				InitialPayment: decimal.NewFromInt(1000000), // 33.33%
				Months:         120,
				Program: model.ProgramRequest{
					Base: true,
				},
			},
			expected: model.Aggregates{
				Rate:            decimal.NewFromFloat(10.0),
				LoanSum:         decimal.NewFromInt(2000000),
				MonthlyPayment:  decimal.NewFromInt(26430),
				Overpayment:     decimal.NewFromInt(1171600),
				LastPaymentDate: "2034-02-18",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calculator.Calculate(tc.request, baseTime)

			// Check for error
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}

			// Check rate
			if !result.Rate.Equal(tc.expected.Rate) {
				t.Errorf("Expected rate %v, got %v", tc.expected.Rate, result.Rate)
			}

			// Check loan sum
			if !result.LoanSum.Equal(tc.expected.LoanSum) {
				t.Errorf("Expected loan sum %v, got %v", tc.expected.LoanSum, result.LoanSum)
			}

			// Check monthly payment
			if !result.MonthlyPayment.Equal(tc.expected.MonthlyPayment) {
				t.Errorf("Expected monthly payment %v, got %v", tc.expected.MonthlyPayment, result.MonthlyPayment)
			}

			// Check overpayment
			if !result.Overpayment.Equal(tc.expected.Overpayment) {
				t.Errorf("Expected overpayment %v, got %v", tc.expected.Overpayment, result.Overpayment)
			}

			// Check last payment date
			if result.LastPaymentDate != tc.expected.LastPaymentDate {
				t.Errorf("Expected last payment date %v, got %v", tc.expected.LastPaymentDate, result.LastPaymentDate)
			}
		})
	}
}

// TestCalculate_Example verifies calculation using the example from the specification
func TestCalculate_Example(t *testing.T) {
	// Create calculator
	calculator := NewMortCalculator()

	// Create request from specification example
	req := model.ExecuteRequest{
		ObjectCost:     decimal.NewFromInt(5000000),
		InitialPayment: decimal.NewFromInt(1000000),
		Months:         240,
		Program: model.ProgramRequest{
			Salary: true,
		},
	}

	// Expected results from specification
	expectedRate := decimal.NewFromFloat(8.0)
	expectedLoanSum := decimal.NewFromInt(4000000)
	expectedMonthlyPayment := decimal.NewFromInt(33458)
	expectedOverpayment := decimal.NewFromInt(4029920)

	// Set fixed loan start date
	baseTime := time.Date(2025, 5, 11, 12, 0, 0, 0, time.UTC)

	// Perform calculation
	result, err := calculator.Calculate(req, baseTime)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check results
	if !result.Rate.Equal(expectedRate) {
		t.Errorf("Expected rate %v, got %v", expectedRate, result.Rate)
	}

	if !result.LoanSum.Equal(expectedLoanSum) {
		t.Errorf("Expected loan sum %v, got %v", expectedLoanSum, result.LoanSum)
	}

	if !result.MonthlyPayment.Equal(expectedMonthlyPayment) {
		t.Errorf("Expected monthly payment %v, got %v", expectedMonthlyPayment, result.MonthlyPayment)
	}

	if !result.Overpayment.Equal(expectedOverpayment) {
		t.Errorf("Expected overpayment %v, got %v", expectedOverpayment, result.Overpayment)
	}

	// Check that the last payment date is 240 months from start
	expectedLastPaymentDate := baseTime.AddDate(0, req.Months, 0).Format(DateFormat)
	if result.LastPaymentDate != expectedLastPaymentDate {
		t.Errorf("Expected last payment date %v, got %v", expectedLastPaymentDate, result.LastPaymentDate)
	}
}

func TestCalculate_ErrorCases(t *testing.T) {
	baseTime := time.Date(2024, 2, 18, 12, 0, 0, 0, time.UTC)
	calculator := NewMortCalculator()

	tests := []struct {
		name        string
		request     model.ExecuteRequest
		expectedErr error
	}{
		{
			name: "Low initial payment",
			request: model.ExecuteRequest{
				ObjectCost:     decimal.NewFromInt(5000000),
				InitialPayment: decimal.NewFromInt(500000), // 10% < 20%
				Months:         240,
				Program: model.ProgramRequest{
					Salary: true,
				},
			},
			expectedErr: model.ErrInitialPaymentLow,
		},
		{
			name: "Zero object cost",
			request: model.ExecuteRequest{
				ObjectCost:     decimal.NewFromInt(0),
				InitialPayment: decimal.NewFromInt(1000000),
				Months:         240,
				Program: model.ProgramRequest{
					Salary: true,
				},
			},
			expectedErr: errors.New("invalid params"),
		},
		{
			name: "Negative number of months",
			request: model.ExecuteRequest{
				ObjectCost:     decimal.NewFromInt(5000000),
				InitialPayment: decimal.NewFromInt(1000000),
				Months:         -10,
				Program: model.ProgramRequest{
					Salary: true,
				},
			},
			expectedErr: errors.New("invalid params"),
		},
		{
			name: "No program selected",
			request: model.ExecuteRequest{
				ObjectCost:     decimal.NewFromInt(5000000),
				InitialPayment: decimal.NewFromInt(1000000),
				Months:         240,
				Program:        model.ProgramRequest{},
			},
			expectedErr: ErrNoProgramSelected,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := calculator.Calculate(tc.request, baseTime)

			// Check that we got the expected error
			if tc.name == "Zero object cost" || tc.name == "Negative number of months" {
				// For the case with "invalid params", we only check the error text
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
				}
			} else {
				if err != tc.expectedErr {
					t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
				}
			}
		})
	}
}

func TestCalculate_Rounding(t *testing.T) {
	baseTime := time.Date(2024, 2, 18, 12, 0, 0, 0, time.UTC)
	calculator := NewMortCalculator()

	// Create a request that should result in non-integer values before rounding
	request := model.ExecuteRequest{
		ObjectCost:     decimal.NewFromInt(5123456), // Amount that should give non-integer values
		InitialPayment: decimal.NewFromInt(1123456),
		Months:         137, // Non-standard number of months
		Program: model.ProgramRequest{
			Salary: true,
		},
	}

	result, err := calculator.Calculate(request, baseTime)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that values are rounded to integers
	rounded := result.MonthlyPayment.Round(0)
	if !result.MonthlyPayment.Equal(rounded) {
		t.Errorf("Monthly payment %v is not rounded to an integer", result.MonthlyPayment)
	}

	rounded = result.Overpayment.Round(0)
	if !result.Overpayment.Equal(rounded) {
		t.Errorf("Overpayment %v is not rounded to an integer", result.Overpayment)
	}

	// Additional check for correct rounding
	loanSum := request.ObjectCost.Sub(request.InitialPayment)
	totalPayment := result.MonthlyPayment.Mul(decimal.NewFromInt(int64(request.Months)))
	calculatedOverpayment := totalPayment.Sub(loanSum)

	// The difference between calculated and returned value should not exceed 1 due to rounding
	diff := calculatedOverpayment.Sub(result.Overpayment).Abs()
	if diff.GreaterThan(decimal.NewFromInt(1)) {
		t.Errorf("Incorrect rounding. Overpayment difference: %v", diff)
	}
}
