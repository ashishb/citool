.PHONY: clean
.PHONY: test

SOURCES=src/main.go src/citool/$(wildcard *.go)
EXECUTABLE=citool

citool: $(SOURCES)
	GO111MODULE=on go build -o $(EXECUTABLE) src/main.go

clean:
	GO111MODULE=on go clean src/main.go
	rm $(EXECUTABLE)

format:
	gofmt -w -s src/*.go src/citool/*

lint:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go vet src/*.go
	GO111MODULE=on go vet src/citool/*
	golint -set_exit_status src/ src/citool/
test:
	./test.sh
