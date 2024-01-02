build:
	@go build -o bin/blockchain

run: build
	@./bin/blockchain

test:
	@go test -v ./... -count=1