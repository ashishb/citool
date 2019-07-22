.PHONY: clean
.PHONY: test

SOURCES=citool/main.go $(wildcard src/citool/*.go)
EXECUTABLE=bin/citool

build: $(SOURCES)
	GO111MODULE=on go build -o $(EXECUTABLE) citool/main.go

clean:
	GO111MODULE=on go clean citool/main.go
	rm -rf $(EXECUTABLE)

format:
	gofmt -w -s citool/main.go citool/src/citool/*

lint:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go vet citool/main.go
	GO111MODULE=on go vet citool/src/citool/*
	golint -set_exit_status citool/src/citool/
test:
	./test.sh
