APPNAME := bin/xkcd-server
DEBUG_APP_NAME := bin/xkcd-d

lint:
	golangci-lint run ./cmd/... ./pkg/... ./internal/...

build: lint
	@echo "Building $(APPNAME)..."
	go build -o $(APPNAME) ./cmd/xkcd/.

d_build:
	go build -gcflags "all=-N -l" -o $(DEBUG_APP_NAME) ./cmd/xkcd/.

podman_up:
	podman build -t comics_db ./build/db/
	podman run  --name comics_db_c -p 5555:5432 -d -v ./.pgdata:/var/lib/postgresql/data:Z  localhost/comics_db:latest 

docker_up:
	docker build -t comics_db ./build/db/
	docker run  --name comics_db_c -p 5555:5432 -d -v ./.pgdata:/var/lib/postgresql/data:Z  comics_db:latest 

server: build podman_up
	./$(APPNAME) -c config.yaml

run: build
	./$(APPNAME) -c config.yaml

all: build

load_test:
	bombardier -n 1000000 -c 125 -l -H "Content-Type: application/json" \
	-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTYyMTcxNTMsInJvbGUiOiJhZG1pbiJ9.5c9Q6G3UlSEvB7SePtAdzIKYjwmaqEBGu1V0JSqIkNI" \
	"http://localhost:4444/pics?search=%22apple%20doctor%22"

clean:
	podman stop comics_db_c || true
	podman rm comics_db_c || true
	podman rmi localhost/comics_db:latest || true
	docker stop comics_db_c || true
	docker rm comics_db_c || true
	docker rmi comics_db || true
