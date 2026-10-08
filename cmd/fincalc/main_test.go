package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantOut  string
		wantFail bool
	}{
		{"loan summary", []string{"loan", "-principal", "25000", "-rate", "0.045", "-months", "60"}, "Monthly payment: $466.08", false},
		{"loan schedule", []string{"loan", "-principal", "1200", "-rate", "0", "-months", "12", "-schedule"}, "   12       100.00", false},
		{"savings projection", []string{"savings", "-start", "100", "-monthly", "50", "-months", "10"}, "Balance after 10 months: $600.00", false},
		{"goal", []string{"goal", "-monthly", "100", "-goal", "1000"}, "Months to reach $1000.00: 10", false},
		{"no command", nil, "", true},
		{"unknown command", []string{"mortgage"}, "", true},
		{"invalid loan input", []string{"loan", "-principal", "0", "-months", "12"}, "", true},
		{"unreachable goal", []string{"goal", "-goal", "100"}, "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := run(tc.args, &buf)
			if tc.wantFail {
				if err == nil {
					t.Fatalf("expected an error, got output %q", buf.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(buf.String(), tc.wantOut) {
				t.Errorf("output missing %q\n got: %s", tc.wantOut, buf.String())
			}
		})
	}
}
