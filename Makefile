# Description: Makefile for healthctl
all: fmt tidy
	go build -o healthctl cmd/main.go

clean: 
	rm -f healthctl

install: 
	cp healthctl /usr/local/bin

uninstall: 
	rm -f /usr/local/bin/healthctl

run: fmt tidy
	go run cmd/main.go

tidy:
	go mod tidy

fmt:
	go fmt ./...


