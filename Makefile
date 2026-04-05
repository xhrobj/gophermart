.PHONY: build run clean

build:
	go build -o cmd/gophermart/ ./cmd/gophermart

run: build
	./cmd/gophermart/gophermart

clean:
	rm -f cmd/gophermart/gophermart
