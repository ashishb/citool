.PHONY: clean
.PHONY: test

build: 
	GO111MODULE=on go build -o citool src/main.go

clean:
	GO111MODULE=on go clean src/main.go
	rm citool

format:
	gofmt -w -s src/*.go src/citool/*

test: 
	./test.sh
