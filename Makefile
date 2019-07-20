.PHONY: clean
.PHONY: test

citool:
	GO111MODULE=on go build -o citool src/main.go

clean:
	GO111MODULE=on go clean src/main.go
	rm citool

format:
	gofmt -w -s src/*.go src/citool/*

lint:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go vet src/*.go
	GO111MODULE=on go vet src/citool/*
	golint -set_exit_status src/ src/citool/
test:
	./test.sh
