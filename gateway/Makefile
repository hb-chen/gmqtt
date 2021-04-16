GOPATH:=$(shell go env GOPATH)

.PHONY: test
test:
	go test -race -cover -v ./...

.PHONY: run
run:
	go run cmd/main.go

.PHONY: build
build:
	CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w' -o ./bin/gateway cmd/main.go

.PHONY: docker
docker: build
	docker build . -t $(tag)
