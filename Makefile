.PHONY: test lint build run stop clean all help

APP_NAME = mortgage-calc
CONTAINER_NAME = mortgage-calc-container
PORT = 8080

test:
	go test -v -cover ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run -c .golangci.yml ./...

build:
	docker build -t $(APP_NAME) .

run:
	docker run -d --name $(CONTAINER_NAME) -p $(PORT):$(PORT) $(APP_NAME)

stop:
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true

clean: stop
	docker rmi $(APP_NAME) || true

all: lint test build
