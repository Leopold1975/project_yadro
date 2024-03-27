APPNAME := bin/myapp

build:
	@echo "Building $(APPNAME)..."
	go build -o $(APPNAME) .

all: build

test: build
	@./$(APPNAME) -s "i'll follow you as long as you are following me" | grep -E 'long' | grep -E 'follow'

bench: 
	go test -bench=. -benchmem > benchmarks.out

.PHONY: build all