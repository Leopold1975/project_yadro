APPNAME := bin/xkcd

build:
	@echo "Building $(APPNAME)..."
	go build -o $(APPNAME) ./cmd/xkcd/.

all: build

test: build
	@./$(APPNAME)
	@./$(APPNAME) -o -n 5 | \
	(grep -oP '"(1|2|3|4|5)":')|| (echo "test failed" && exit 1)
	go test -count=1 -v ./...
	@echo tests ok
