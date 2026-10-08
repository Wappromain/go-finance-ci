package loan

import (
	"errors"
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestMonthlyPayment(t *testing.T) {
	tests := []struct {
		name       string
		principal  float64
		annualRate float64
		months     int
		want       float64
	}{
		{"standard 30 year mortgage", 200000, 0.06, 360, 1199.10},
		{"five year car loan", 25000, 0.045, 60, 466.08},
		{"small microloan", 1000, 0.12, 12, 88.85},
		{"zero interest splits evenly", 1200, 0, 12, 100.00},
		{"single payment", 500, 0.12, 1, 505.00},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MonthlyPayment(tc.principal, tc.annualRate, tc.months)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !almostEqual(got, tc.want) {
				t.Errorf("MonthlyPayment(%v, %v, %v) = %.2f, want %.2f",
					tc.principal, tc.annualRate, tc.months, got, tc.want)
			}
		})
	}
}

func TestMonthlyPaymentRejectsBadInput(t *testing.T) {
	tests := []struct {
		name       string
		principal  float64
		annualRate float64
		months     int
		wantErr    error
	}{
		{"zero principal", 0, 0.05, 12, ErrInvalidPrincipal},
		{"negative principal", -100, 0.05, 12, ErrInvalidPrincipal},
		{"negative rate", 1000, -0.01, 12, ErrInvalidRate},
		{"zero months", 1000, 0.05, 0, ErrInvalidTerm},
		{"negative months", 1000, 0.05, -6, ErrInvalidTerm},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := MonthlyPayment(tc.principal, tc.annualRate, tc.months)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got error %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestScheduleLength(t *testing.T) {
	sched, err := Schedule(10000, 0.05, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sched) != 24 {
		t.Fatalf("len(schedule) = %d, want 24", len(sched))
	}
}

func TestScheduleEndsAtZeroBalance(t *testing.T) {
	sched, err := Schedule(10000, 0.05, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	last := sched[len(sched)-1]
	if !almostEqual(last.Balance, 0) {
		t.Errorf("final balance = %.2f, want 0.00", last.Balance)
	}
}

func TestSchedulePrincipalSumsToLoan(t *testing.T) {
	const principal = 10000.0
	sched, err := Schedule(principal, 0.05, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var paid float64
	for _, p := range sched {
		paid += p.Principal
	}
	if !almostEqual(paid, principal) {
		t.Errorf("sum of principal = %.2f, want %.2f", paid, principal)
	}
}

func TestScheduleInterestDecreasesOverTime(t *testing.T) {
	sched, err := Schedule(10000, 0.05, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(sched); i++ {
		if sched[i].Interest > sched[i-1].Interest {
			t.Fatalf("interest rose from month %d (%.2f) to month %d (%.2f)",
				i, sched[i-1].Interest, i+1, sched[i].Interest)
		}
	}
}

func TestTotalInterest(t *testing.T) {
	tests := []struct {
		name       string
		principal  float64
		annualRate float64
		months     int
		want       float64
	}{
		{"car loan", 25000, 0.045, 60, 2964.94},
		{"no interest", 1200, 0, 12, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := TotalInterest(tc.principal, tc.annualRate, tc.months)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tc.want) > 0.05 {
				t.Errorf("TotalInterest = %.2f, want %.2f", got, tc.want)
			}
		})
	}
}
