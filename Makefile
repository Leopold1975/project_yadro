APPNAME := bin/xkcd-server
DEBUG_APP_NAME := /bin/xkcd-d

build:
	@echo "Building $(APPNAME)..."
	go build -o $(APPNAME) ./cmd/xkcd/.

d_build:
	go build -gcflags "all=-N -l" -o $(DEBUG_APP_NAME) ./cmd/xkcd/.

server: build
	./$(APPNAME) -c config.yaml

all: build
