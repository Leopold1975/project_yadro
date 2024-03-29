APPNAME := bin/myapp

build:
	@echo "Building $(APPNAME)..."
	go build -o $(APPNAME) .

all: build

test: build
	@./$(APPNAME) -s "i'll follow you as long as you are following me" | grep -E 'long' | grep -E 'follow'
	go test -count=1 -v .

bench: 
	go test -bench=. -benchmem > benchmarks.out
