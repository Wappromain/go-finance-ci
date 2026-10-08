// Command fincalc is a small command line front end for the loan and
// savings packages.
//
// Usage:
//
//	fincalc loan    -principal 25000 -rate 0.045 -months 60 [-schedule]
//	fincalc savings -start 500 -monthly 100 -rate 0.05 -months 24
//	fincalc goal    -start 0 -monthly 100 -rate 0.06 -goal 5000
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/patrick/go-finance-ci/loan"
	"github.com/patrick/go-finance-ci/savings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

const usage = `usage: fincalc <loan|savings|goal> [flags]
run "fincalc <command> -h" for the flags of each command`

// run is separated from main so it can be tested without exiting the process.
func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", usage)
	}
	switch args[0] {
	case "loan":
		return runLoan(args[1:], out)
	case "savings":
		return runSavings(args[1:], out)
	case "goal":
		return runGoal(args[1:], out)
	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], usage)
	}
}

func runLoan(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("loan", flag.ContinueOnError)
	fs.SetOutput(out)
	principal := fs.Float64("principal", 0, "amount borrowed")
	rate := fs.Float64("rate", 0, "annual interest rate as a decimal (0.05 = 5%)")
	months := fs.Int("months", 0, "loan term in months")
	schedule := fs.Bool("schedule", false, "print the full amortization schedule")
	if err := fs.Parse(args); err != nil {
		return err
	}

	pmt, err := loan.MonthlyPayment(*principal, *rate, *months)
	if err != nil {
		return err
	}
	interest, err := loan.TotalInterest(*principal, *rate, *months)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Monthly payment: $%.2f\n", pmt)
	fmt.Fprintf(out, "Total interest:  $%.2f\n", interest)
	fmt.Fprintf(out, "Total paid:      $%.2f\n", *principal+interest)

	if *schedule {
		rows, err := loan.Schedule(*principal, *rate, *months)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "\n%5s %12s %12s %12s %14s\n", "Month", "Payment", "Principal", "Interest", "Balance")
		for _, r := range rows {
			fmt.Fprintf(out, "%5d %12.2f %12.2f %12.2f %14.2f\n",
				r.Month, r.Payment, r.Principal, r.Interest, r.Balance)
		}
	}
	return nil
}

func runSavings(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("savings", flag.ContinueOnError)
	fs.SetOutput(out)
	start := fs.Float64("start", 0, "starting balance")
	monthly := fs.Float64("monthly", 0, "deposit added each month")
	rate := fs.Float64("rate", 0, "annual interest rate as a decimal")
	months := fs.Int("months", 12, "number of months to project")
	if err := fs.Parse(args); err != nil {
		return err
	}
	fv, err := savings.FutureValue(*start, *monthly, *rate, *months)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Balance after %d months: $%.2f\n", *months, fv)
	return nil
}

func runGoal(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("goal", flag.ContinueOnError)
	fs.SetOutput(out)
	start := fs.Float64("start", 0, "starting balance")
	monthly := fs.Float64("monthly", 0, "deposit added each month")
	rate := fs.Float64("rate", 0, "annual interest rate as a decimal")
	goal := fs.Float64("goal", 0, "target balance")
	if err := fs.Parse(args); err != nil {
		return err
	}
	m, err := savings.MonthsToGoal(*start, *monthly, *rate, *goal)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Months to reach $%.2f: %d\n", *goal, m)
	return nil
}
