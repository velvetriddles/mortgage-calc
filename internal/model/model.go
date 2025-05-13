package model

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrChooseNone        = errors.New("no mortgage program selected")
	ErrChooseMultiple    = errors.New("multiple mortgage programs selected")
	ErrInitialPaymentLow = errors.New("initial payment is too low")
)

type ProgramRequest struct {
	Salary   bool `json:"salary"`
	Military bool `json:"military"`
	Base     bool `json:"base"`
}

type RequestParams struct {
	ObjectCost     decimal.Decimal `json:"object_cost"`
	InitialPayment decimal.Decimal `json:"initial_payment"`
	Months         int             `json:"months"`
}

type ExecuteRequest struct {
	ObjectCost     decimal.Decimal `json:"object_cost"`
	InitialPayment decimal.Decimal `json:"initial_payment"`
	Months         int             `json:"months"`
	Program        ProgramRequest  `json:"program"`
}

type Aggregates struct {
	Rate            decimal.Decimal `json:"rate"`
	LoanSum         decimal.Decimal `json:"loan_sum"`
	MonthlyPayment  decimal.Decimal `json:"monthly_payment"`
	Overpayment     decimal.Decimal `json:"overpayment"`
	LastPaymentDate string          `json:"last_payment_date"`
}

type ExecuteResponse struct {
	Params     RequestParams  `json:"params"`
	Program    ProgramRequest `json:"program"`
	Aggregates Aggregates     `json:"aggregates"`
}
