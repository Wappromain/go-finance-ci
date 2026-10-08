.PHONY: all fmt vet test cover build run clean

BINARY := bin/fincalc

all: fmt vet test build

fmt:
	gofmt -l -w .

vet:
	go vet ./...

test:
	go test -race -count=1 ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

build:
	go build -o $(BINARY) ./cmd/fincalc

run: build
	./$(BINARY) loan -principal 25000 -rate 0.045 -months 60

clean:
	rm -rf bin coverage.out
