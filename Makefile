GOPATH:=$(shell go env GOPATH)

.PHONY: test
test:
	go test -race -cover -v ./...

.PHONY: run
run:
	go run cmd/main.go

.PHONY: gateway
gateway:
	CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w' -o ./bin/gateway cmd/gateway/main.go

.PHONY: console
console:
	CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w' -o ./bin/console cmd/console/main.go

.PHONY: api
api:
	CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w' -o ./bin/api cmd/api/main.go

.PHONY: docker
docker: build
	docker build . -t $(tag)

.PHONY: release
release:
	goreleaser release --config .goreleaser.yml --skip-validate --skip-publish --rm-dist

.PHONY: snapshot
snapshot:
	goreleaser release --config .goreleaser.yml --skip-publish --snapshot --rm-dist
