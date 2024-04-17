APPNAME := bin/xkcd

build:
	@echo "Building $(APPNAME)..."
	go build -o $(APPNAME) ./cmd/xkcd/.

all: build
