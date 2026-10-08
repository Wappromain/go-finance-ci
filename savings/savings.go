// Package savings projects the growth of a savings balance with monthly
// deposits and monthly compounding interest.
package savings

import (
	"errors"
	"math"
)

// Errors returned for invalid savings parameters.
var (
	ErrNegativeAmount = errors.New("savings: amounts cannot be negative")
	ErrNegativeRate   = errors.New("savings: annual rate cannot be negative")
	ErrNegativeMonths = errors.New("savings: months cannot be negative")
	ErrUnreachable    = errors.New("savings: goal can never be reached with these inputs")
)

// maxMonths caps MonthsToGoal so very slow plans terminate (100 years).
const maxMonths = 1200

// FutureValue returns the balance after the given number of months, starting
// from start, adding monthly at the end of each month, with interest
// compounded monthly at annualRate (0.05 means 5%).
func FutureValue(start, monthly, annualRate float64, months int) (float64, error) {
	switch {
	case start < 0 || monthly < 0:
		return 0, ErrNegativeAmount
	case annualRate < 0:
		return 0, ErrNegativeRate
	case months < 0:
		return 0, ErrNegativeMonths
	}
	r := annualRate / 12
	n := float64(months)
	if r == 0 {
		return start + monthly*n, nil
	}
	growth := math.Pow(1+r, n)
	return start*growth + monthly*(growth-1)/r, nil
}

// MonthsToGoal returns how many whole months it takes for the balance to
// reach goal. It returns ErrUnreachable if the goal is not met within 100 years.
func MonthsToGoal(start, monthly, annualRate, goal float64) (int, error) {
	if start < 0 || monthly < 0 || goal < 0 {
		return 0, ErrNegativeAmount
	}
	if annualRate < 0 {
		return 0, ErrNegativeRate
	}
	r := annualRate / 12
	balance := start
	for m := 0; m <= maxMonths; m++ {
		if balance >= goal {
			return m, nil
		}
		balance = balance*(1+r) + monthly
	}
	return 0, ErrUnreachable
}
