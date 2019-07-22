.PHONY: clean
.PHONY: test

SOURCES=citool.go $(wildcard src/citool/*.go)
EXECUTABLE=citool

citool: $(SOURCES)
	GO111MODULE=on go build -o bin/$(EXECUTABLE) citool.go

clean:
	GO111MODULE=on go clean citool.go
	rm -rf $(EXECUTABLE)

format:
	gofmt -w -s citool.go src/citool/*

lint:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go vet citool.go
	GO111MODULE=on go vet src/citool/*
	golint -set_exit_status src/ src/citool/
test:
	./test.sh

all:
	make clean
	make
	make lint
	make test
