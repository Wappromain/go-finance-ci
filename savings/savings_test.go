package savings

import (
	"errors"
	"math"
	"testing"
)

func TestFutureValue(t *testing.T) {
	tests := []struct {
		name       string
		start      float64
		monthly    float64
		annualRate float64
		months     int
		want       float64
	}{
		{"no interest just adds deposits", 100, 50, 0, 10, 600},
		{"lump sum only, one year at 12%", 1000, 0, 0.12, 12, 1126.83},
		{"deposits only, one year at 6%", 0, 100, 0.06, 12, 1233.56},
		{"zero months returns start", 500, 100, 0.05, 0, 500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FutureValue(tc.start, tc.monthly, tc.annualRate, tc.months)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tc.want) > 0.01 {
				t.Errorf("FutureValue = %.2f, want %.2f", got, tc.want)
			}
		})
	}
}

func TestFutureValueRejectsNegatives(t *testing.T) {
	if _, err := FutureValue(-1, 0, 0.05, 12); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("negative start: got %v, want ErrNegativeAmount", err)
	}
	if _, err := FutureValue(0, -1, 0.05, 12); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("negative deposit: got %v, want ErrNegativeAmount", err)
	}
	if _, err := FutureValue(0, 10, -0.05, 12); !errors.Is(err, ErrNegativeRate) {
		t.Errorf("negative rate: got %v, want ErrNegativeRate", err)
	}
	if _, err := FutureValue(0, 10, 0.05, -1); !errors.Is(err, ErrNegativeMonths) {
		t.Errorf("negative months: got %v, want ErrNegativeMonths", err)
	}
}

func TestMonthsToGoal(t *testing.T) {
	tests := []struct {
		name       string
		start      float64
		monthly    float64
		annualRate float64
		goal       float64
		want       int
	}{
		{"already reached", 1000, 0, 0.05, 500, 0},
		{"simple deposits, no interest", 0, 100, 0, 1000, 10},
		{"rounds up partial month", 0, 300, 0, 1000, 4},
		{"interest shortens the wait", 0, 100, 0.06, 1233, 12},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MonthsToGoal(tc.start, tc.monthly, tc.annualRate, tc.goal)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("MonthsToGoal = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestMonthsToGoalUnreachable(t *testing.T) {
	_, err := MonthsToGoal(0, 0, 0, 100)
	if !errors.Is(err, ErrUnreachable) {
		t.Errorf("got %v, want ErrUnreachable", err)
	}
}
