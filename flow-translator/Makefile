VERSION= 0.1
LDFLAGS= -ldflags "-X main.version=${VERSION}"

deps:
	dep ensure

build:
	cd flow-translator; go build $(LDFLAGS) -o flow-translator

run: build
	cd flow-translator; ./flow-translator kafka-consumer --config-file ../flow-translator.conf

test:
	go test -v ./...
