# go-finance-ci

![CI](https://github.com/<Wappromain>/go-finance-ci/actions/workflows/ci.yml/badge.svg)

A small Go library and command line tool for personal finance calculations (loan payments, amortization schedules, savings projections), built with test driven development and checked by a GitHub Actions workflow on every change. It is the foundation repository for this semester's assignments.

## What the code does

| Package | Function | Purpose |
|---|---|---|
| `loan` | `MonthlyPayment` | Fixed monthly payment for an amortizing loan |
| `loan` | `Schedule` | Month by month split of principal, interest and remaining balance |
| `loan` | `TotalInterest` | Total interest paid over the life of the loan |
| `savings` | `FutureValue` | Balance after N months of deposits with monthly compounding |
| `savings` | `MonthsToGoal` | Months needed to reach a savings target |
| `cmd/fincalc` | `fincalc` | Command line front end for all of the above |

Every function validates its input and returns a typed error (for example `loan.ErrInvalidPrincipal`) instead of producing nonsense numbers.

## Repository layout

```
go-finance-ci/
├── .github/workflows/ci.yml   # CI pipeline
├── cmd/fincalc/               # CLI (main.go, main_test.go)
├── loan/                      # loan package and tests
├── savings/                   # savings package and tests
├── go.mod                     # module and Go version (dependency management)
├── Makefile                   # shortcuts for common commands
└── README.md
```

## Prerequisites

* Go 1.24 or newer ([download](https://go.dev/dl/)). Check with `go version`.
* Git
* `make` (optional; every Make target has a plain `go` equivalent below)

## Getting started

```bash
git clone https://github.com/<Wappromain>/go-finance-ci.git
cd go-finance-ci
go mod download        # fetch dependencies (none yet beyond the standard library)
go test ./...          # run every test
```

## Running the tests

| Task | With Make | Plain Go |
|---|---|---|
| Run all tests (with race detector) | `make test` | `go test -race -count=1 ./...` |
| Verbose output | | `go test -v ./...` |
| One package | | `go test ./loan` |
| One test | | `go test ./loan -run TestMonthlyPayment` |
| Coverage report | `make cover` | `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out` |
| Coverage in browser | | `go tool cover -html=coverage.out` |
| Format and static checks | `make fmt vet` | `gofmt -l -w . && go vet ./...` |
| Everything CI runs | `make all` | |

Note: the `-race` flag needs cgo. If it errors on your machine, drop it.

## Running the program

Build the binary:

```bash
make build                         # or: go build -o bin/fincalc ./cmd/fincalc
```

Or run without building using `go run ./cmd/fincalc ...`.

Rates are decimals: `0.045` means 4.5%.

```bash
# Monthly payment for a $25,000, 5 year car loan at 4.5%
./bin/fincalc loan -principal 25000 -rate 0.045 -months 60
# Monthly payment: $466.08
# Total interest:  $2964.53
# Total paid:      $27964.53

# Add -schedule to print the full amortization table
./bin/fincalc loan -principal 1200 -rate 0.06 -months 3 -schedule

# Project savings: $500 to start, $100 a month, 5% for 2 years
./bin/fincalc savings -start 500 -monthly 100 -rate 0.05 -months 24
# Balance after 24 months: $3071.06

# How long to save $5,000 at $150 a month and 4%?
./bin/fincalc goal -monthly 150 -rate 0.04 -goal 5000
# Months to reach $5000.00: 32
```

Use `./bin/fincalc <command> -h` to list the flags for any command. Invalid input exits with status 1 and an error message.

## Continuous integration

The workflow in `.github/workflows/ci.yml` runs automatically when anyone pushes or opens a pull request that touches Go code, tests, `go.mod`/`go.sum`, the Makefile, or the workflow itself. Changes to documentation alone do not trigger it.

To run it on demand: open the **Actions** tab, select **CI**, then **Run workflow**.

The pipeline has three stages. Each must pass before the next one starts, so a failure stops the pipeline and reports back on the commit or pull request.

1. **Format and vet**: confirms `go.mod` is tidy, every file is `gofmt` formatted, and `go vet` finds no suspicious code.
2. **Test**: runs the full suite with the race detector on Linux, macOS and Windows, against both the minimum supported Go version (1.24) and the latest stable release. The Linux stable run posts a coverage table to the run summary.
3. **Build**: compiles `fincalc`, smoke tests it with a real command, and uploads the binary as a downloadable artifact (kept 7 days). This is the continuous delivery step: every passing commit produces a release ready binary.

Recommended repository setting: under **Settings → Branches**, add a protection rule on `main` that requires the CI checks to pass before merging.

## Test driven development workflow

The commit history shows the cycle this project follows:

1. **Red**: write a test describing the new behavior, run `go test ./...`, and watch it fail.
2. **Green**: write the smallest amount of code that makes it pass.
3. **Refactor**: clean up with the tests as a safety net, then push and let CI confirm.

Tests are table driven: each test function holds a slice of named cases, so adding a scenario is one line. For example, to cover a new loan case, add a row to `TestMonthlyPayment` in `loan/loan_test.go`:

```go
{"student loan", 30000, 0.055, 120, 325.58},
```

The first commit in this repository contains only tests (red); the second adds the implementation (green). During that step one expected value turned out to be wrong because it had been hand calculated from a rounded payment. The test was corrected rather than the code, which is a normal part of TDD: the tests are a specification and are checked as carefully as the code.

## License

MIT
