package service

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/velvetriddles/mortgage-calc/internal/model"
)

var (
	// Annual interest rates for different credit programs
	RateSalaryProgram   = decimal.NewFromFloat(8.0)
	RateMilitaryProgram = decimal.NewFromFloat(9.0)
	RateBaseProgram     = decimal.NewFromFloat(10.0)

	// Minimum percentage of initial payment
	MinInitialPaymentPercent = decimal.NewFromFloat(0.2)

	// Date format for the last payment
	DateFormat = "2006-01-02"

	// Constants for mathematical calculations
	DecimalZero    = decimal.NewFromInt(0)
	DecimalOne     = decimal.NewFromInt(1)
	DecimalTwelve  = decimal.NewFromInt(12)
	DecimalHundred = decimal.NewFromInt(100)

	// Error for when no program is selected
	ErrNoProgramSelected = errors.New("no mortgage program selected")
)

// Calculator defines the interface for mortgage calculations
type Calculator interface {
	Calculate(req model.ExecuteRequest, baseTime time.Time) (model.Aggregates, error)
}

// MortCalculator implements mortgage parameter calculations
type MortCalculator struct{}

// NewMortCalculator creates a new instance of the mortgage calculator
func NewMortCalculator() *MortCalculator {
	return &MortCalculator{}
}

// Calculate performs mortgage calculation based on input data
// baseTime is used as the base date for calculating the last payment date (for testing)
func (c *MortCalculator) Calculate(req model.ExecuteRequest, baseTime time.Time) (model.Aggregates, error) {
	if req.ObjectCost.LessThanOrEqual(DecimalZero) || req.Months <= 0 {
		return model.Aggregates{}, errors.New("invalid params")
	}

	rate, err := c.getProgramRate(req.Program)
	if err != nil {
		return model.Aggregates{}, err
	}

	minPayment := req.ObjectCost.Mul(MinInitialPaymentPercent)
	if req.InitialPayment.LessThan(minPayment) {
		return model.Aggregates{}, model.ErrInitialPaymentLow
	}

	currentTime := baseTime
	if currentTime.IsZero() {
		currentTime = time.Now()
	}

	loanSum := req.ObjectCost.Sub(req.InitialPayment)

	// Calculate monthly payment using annuity formula:
	// P = (S * r * (1 + r)^n) / ((1 + r)^n - 1)
	// where:
	// P - monthly payment
	// S - loan amount
	// r - monthly interest rate (annual rate / 12)
	// n - number of months (loan term)

	monthlyRate := rate.Div(DecimalHundred).Div(DecimalTwelve)

	// (1 + r)^n
	power := DecimalOne.Add(monthlyRate).Pow(decimal.NewFromInt(int64(req.Months)))

	// r * (1 + r)^n / ((1 + r)^n - 1)
	annuityCoeff := monthlyRate.Mul(power).Div(power.Sub(DecimalOne))

	monthlyPayment := loanSum.Mul(annuityCoeff).Round(0)

	totalPayment := monthlyPayment.Mul(decimal.NewFromInt(int64(req.Months)))
	overpayment := totalPayment.Sub(loanSum).Round(0)

	lastPaymentDate := currentTime.AddDate(0, req.Months, 0)

	return model.Aggregates{
		Rate:            rate,
		LoanSum:         loanSum,
		MonthlyPayment:  monthlyPayment,
		Overpayment:     overpayment,
		LastPaymentDate: lastPaymentDate.Format(DateFormat),
	}, nil
}

// getProgramRate returns the interest rate based on the selected program
func (c *MortCalculator) getProgramRate(program model.ProgramRequest) (decimal.Decimal, error) {
	switch {
	case program.Salary:
		return RateSalaryProgram, nil
	case program.Military:
		return RateMilitaryProgram, nil
	case program.Base:
		return RateBaseProgram, nil
	default:
		return DecimalZero, ErrNoProgramSelected
	}
}
