// Package loan calculates fixed rate, fully amortizing loan payments
// and produces month by month repayment schedules.
package loan

import (
	"errors"
	"math"
)

// Errors returned for invalid loan parameters.
var (
	ErrInvalidPrincipal = errors.New("loan: principal must be greater than zero")
	ErrInvalidRate      = errors.New("loan: annual rate cannot be negative")
	ErrInvalidTerm      = errors.New("loan: term must be at least one month")
)

// Payment is one row of an amortization schedule.
type Payment struct {
	Month     int
	Payment   float64
	Principal float64
	Interest  float64
	Balance   float64
}

func validate(principal, annualRate float64, months int) error {
	switch {
	case principal <= 0:
		return ErrInvalidPrincipal
	case annualRate < 0:
		return ErrInvalidRate
	case months < 1:
		return ErrInvalidTerm
	}
	return nil
}

// MonthlyPayment returns the fixed monthly payment for a loan of the given
// principal, annual interest rate (0.05 means 5%) and term in months.
func MonthlyPayment(principal, annualRate float64, months int) (float64, error) {
	if err := validate(principal, annualRate, months); err != nil {
		return 0, err
	}
	r := annualRate / 12
	if r == 0 {
		return principal / float64(months), nil
	}
	growth := math.Pow(1+r, float64(months))
	return principal * r * growth / (growth - 1), nil
}

// Schedule returns the full amortization schedule for the loan. The final
// payment absorbs any rounding so the balance ends at exactly zero.
func Schedule(principal, annualRate float64, months int) ([]Payment, error) {
	pmt, err := MonthlyPayment(principal, annualRate, months)
	if err != nil {
		return nil, err
	}
	r := annualRate / 12
	balance := principal
	rows := make([]Payment, 0, months)

	for m := 1; m <= months; m++ {
		interest := balance * r
		toPrincipal := pmt - interest
		if m == months {
			toPrincipal = balance
		}
		balance -= toPrincipal
		rows = append(rows, Payment{
			Month:     m,
			Payment:   toPrincipal + interest,
			Principal: toPrincipal,
			Interest:  interest,
			Balance:   math.Max(balance, 0),
		})
	}
	return rows, nil
}

// TotalInterest returns the total interest paid over the life of the loan.
func TotalInterest(principal, annualRate float64, months int) (float64, error) {
	sched, err := Schedule(principal, annualRate, months)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, p := range sched {
		total += p.Interest
	}
	return total, nil
}
