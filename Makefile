# Description: Makefile for healthctl

all:
	go build -o healthctl cmd/main.go

clean:
	rm -f healthctl

install:
	cp healthctl /usr/local/bin

uninstall:
	rm -f /usr/local/bin/healthctl

run:
	go run cmd/main.go


