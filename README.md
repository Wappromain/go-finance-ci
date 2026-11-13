# go-finance-ci

![CI](https://github.com/<your-username>/go-finance-ci/actions/workflows/ci.yml/badge.svg)
![API tests](https://github.com/<your-username>/go-finance-ci/actions/workflows/api.yml/badge.svg)

A Go library, command line tool and REST API for personal finance calculations (loan payments, amortization schedules, savings projections). It is built with test driven development, packaged with Docker, and checked by GitHub Actions on every change, with results posted to Slack. It is the foundation repository for this semester's assignments.

## Contents

* [Repository layout](#repository-layout)
* [Prerequisites](#prerequisites)
* [REST API](#rest-api)
* [Running the API and tests with Docker](#running-the-api-and-tests-with-docker)
* [Running without Docker](#running-without-docker)
* [Continuous integration](#continuous-integration)
* [Slack notifications](#slack-notifications)
* [Command line calculator](#command-line-calculator)
* [Test driven development workflow](#test-driven-development-workflow)

## Repository layout

```
go-finance-ci/
├── .github/workflows/
│   ├── ci.yml              # lint, test matrix, CLI build (assignment 1)
│   └── api.yml             # Docker tests, API image check, Slack summary
├── api/                    # REST API handlers and their tests
├── cmd/
│   ├── server/             # starts the REST API
│   └── fincalc/            # command line calculator
├── loan/                   # loan calculations and tests
├── savings/                # savings calculations and tests
├── scripts/
│   ├── run_api.sh          # build + run the API container until stopped
│   ├── run_tests.sh        # build + run the test container, exit 0/non zero
│   └── notify_slack.sh     # post a workflow summary to Slack
├── Dockerfile.api          # image that runs the API
├── Dockerfile.test         # image that runs the tests
├── go.mod                  # module and Go version (dependency management)
└── Makefile
```

## Prerequisites

* [Docker](https://docs.docker.com/get-docker/) (Docker Desktop on Windows or macOS). This is all you need to run the API and the tests.
* Git
* Optional, to run without Docker: Go 1.24 or newer
* On Windows, run the scripts from Git Bash or WSL.

## REST API

The API manages loans in memory. Every saved loan includes its monthly payment, calculated by the `loan` package and rounded to cents. Data resets when the server restarts.

### Endpoints

| Method | Path | Success | Errors |
|---|---|---|---|
| `GET` | `/loans` | `200` list of all loans | none |
| `POST` | `/loans` | `201` the created loan | `400` invalid JSON or values |
| `GET` | `/loans/{id}` | `200` the loan | `400` id not a number, `404` not found |
| `PUT` | `/loans/{id}` | `200` the replaced loan | `400` bad id, JSON or values, `404` not found |
| `DELETE` | `/loans/{id}` | `204` no body | `400` id not a number, `404` not found |

Any other method on these paths returns `405 Method Not Allowed`.

### Loan fields

| Field | Type | Sent by client | Rules |
|---|---|---|---|
| `id` | integer | no, assigned by the server | |
| `principal` | number | yes | greater than 0 |
| `rate` | number | yes | annual rate as a decimal (`0.045` = 4.5%), not negative |
| `months` | integer | yes | at least 1 |
| `monthly_payment` | number | no, calculated | |

Errors return JSON such as `{"error": "loan not found"}`.

### Example session

With the API running on port 8080:

```bash
# Create (201)
curl -i -X POST localhost:8080/loans \
  -d '{"principal": 25000, "rate": 0.045, "months": 60}'
# {"id":1,"principal":25000,"rate":0.045,"months":60,"monthly_payment":466.08}

# Read all (200), read one (200)
curl -i localhost:8080/loans
curl -i localhost:8080/loans/1

# Replace (200)
curl -i -X PUT localhost:8080/loans/1 \
  -d '{"principal": 1200, "rate": 0, "months": 12}'
# {"id":1,"principal":1200,"rate":0,"months":12,"monthly_payment":100}

# Delete (204), then it is gone (404)
curl -i -X DELETE localhost:8080/loans/1
curl -i localhost:8080/loans/1
```

`-i` prints the status line, so you can see each status code.

## Running the API and tests with Docker

Each script builds its image and runs it, so a single command does everything. Run them from the repository root.

### Run the API (until you stop it)

```bash
./scripts/run_api.sh              # http://localhost:8080
PORT=9000 ./scripts/run_api.sh    # http://localhost:9000
```

Press **Ctrl+C** to stop. The server shuts down cleanly and the container is removed.

Equivalent manual commands:

```bash
docker build -f Dockerfile.api -t loan-api .
docker run --rm -p 8080:8080 loan-api
```

`Dockerfile.api` is a two stage build. A full Go image compiles a static binary, which is then copied into a minimal distroless image that runs as a non root user, with no shell and no compiler.

### Run the tests (exits when done)

```bash
./scripts/run_tests.sh
echo $?     # 0 = all tests passed, non zero = at least one failed
```

Equivalent manual commands:

```bash
docker build -f Dockerfile.test -t loan-api-test .
docker run --rm loan-api-test
```

`Dockerfile.test` runs `go test -v -race ./...` as its command. `go test` exits non zero if any test fails, `docker run` passes that exit status through, and the script (with `set -e`) returns it as its own.

If you get `permission denied` running a script, either run it with `bash scripts/run_tests.sh` or make it executable with `chmod +x scripts/*.sh`.

## Running without Docker

Requires Go 1.24+.

```bash
go run ./cmd/server                 # API on :8080 (PORT=9000 to change)
go test ./...                       # all tests
go test ./api -v                    # API tests only, verbose
go test ./api -run TestDeleteLoan   # one group of tests
make all                            # format, vet, test, build the CLI
```

## Continuous integration

There are two workflows. Both run on pushes and pull requests that change relevant files, and both can be started manually from **Actions → (workflow) → Run workflow**.

### `api.yml`: Docker tests and Slack

Runs when Go code, `go.mod`/`go.sum`, either Dockerfile, anything in `scripts/`, or the workflow itself changes.

| Job | What it does |
|---|---|
| **Run tests in Docker** | Runs `scripts/run_tests.sh`, the same command you run locally, so CI tests exactly what you test. It counts passing and failing tests and writes a results table to the run summary. |
| **Build and smoke test API image** | Builds `Dockerfile.api`, starts the container, and sends a real request with every verb, checking each status code (201, 200, 200, 200, 204, 404). |
| **Post summary to Slack** | Runs after both jobs, **whether they passed or failed**, and posts the overall status to Slack. |

### `ci.yml`: library checks

Checks formatting and `go vet`, runs the tests on Linux, macOS and Windows against Go 1.24 and the latest stable release, and builds the CLI binary.

## Slack notifications

After each `api.yml` run, a message like this is posted:

> ✅ **API tests passed**
> **Repository** patrick/go-finance-ci  **Branch** main
> **Commit** a1b2c3d feat(api): add PUT and DELETE
> **Triggered by** patrick (push)
> **Tests (Docker)** ✅ success (24 passed, 0 failed)  **API image** ✅ success
> [View workflow run]

### Setup (one time)

1. **Create a Slack app with an incoming webhook**
   * Go to <https://api.slack.com/apps>, choose **Create New App → From scratch**, give it a name (for example *CI Bot*) and pick your workspace.
   * Open **Incoming Webhooks**, switch it **On**, then click **Add New Webhook to Workspace**, choose the channel, and click **Allow**.
   * Copy the webhook URL (`https://hooks.slack.com/services/...`).
2. **Store it as a GitHub secret.** Never commit it, because anyone with the URL can post to your channel.
   * Repository **Settings → Secrets and variables → Actions → New repository secret**
   * Name: `SLACK_WEBHOOK_URL`, Value: the URL.
3. **Trigger a run** from Actions → API tests → Run workflow, and check the channel.

If the secret is missing, the notify job logs a warning and succeeds, so forks and new clones don't fail.

### Test the message from your machine

```bash
SLACK_WEBHOOK_URL='https://hooks.slack.com/services/...' ./scripts/notify_slack.sh
```

The script uses `curl` and `jq`, and builds the message JSON with `jq` so quotes or symbols in commit messages can't break it.

## Command line calculator

```bash
go build -o bin/fincalc ./cmd/fincalc
./bin/fincalc loan -principal 25000 -rate 0.045 -months 60 [-schedule]
./bin/fincalc savings -start 500 -monthly 100 -rate 0.05 -months 24
./bin/fincalc goal -monthly 150 -rate 0.04 -goal 5000
```

## Test driven development workflow

Each feature was written as **red → green → refactor**:

1. **Red:** write a test for the new behavior and watch it fail.
2. **Green:** write the smallest amount of code that makes it pass.
3. **Refactor:** clean up with the tests as a safety net, then push and let CI confirm.

The API was built one verb at a time in this order: `GET /loans`, `POST`, `GET /loans/{id}`, then `PUT` and `DELETE`. Each verb's tests were committed before or with its implementation. Tests use `net/http/httptest`, which sends requests straight to the handler without starting a real server, so the full suite runs in milliseconds.

Most tests are table driven: each test function holds a list of named cases, so a new scenario is one new row. For example, to check that a negative principal is rejected on update, add this row to `TestUpdateLoanErrors` in `api/update_delete_test.go`:

```go
{"negative principal", "/loans/1", `{"principal": -5, "rate": 0.05, "months": 12}`, http.StatusBadRequest},
```

## License

MIT
