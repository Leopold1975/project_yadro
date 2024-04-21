APPNAME := bin/xkcd

build:
	@echo "Building $(APPNAME)..."
	go build -o $(APPNAME) ./cmd/xkcd/.

dbuild:
	go build -gcflags "all=-N -l" -o $(APPNAME) ./cmd/xkcd/.

run:
	./bin/xkcd -c config.yaml

all: build
