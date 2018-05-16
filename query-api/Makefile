VERSION= 0.1
LDFLAGS= -ldflags "-X main.version=${VERSION}"

deps:
	dep ensure

build:
	cd query-api; go build $(LDFLAGS) -o query-api

run: build
	cd query-api; ./query-api --config-file ../query-api.conf

test:
	go test -v ./...
