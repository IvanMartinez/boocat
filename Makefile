# Targets
#
.PHONY: debug
debug:	### Debug Makefile itself
	@echo $(UNAME)

.PHONY: build
build:
	go build -o strki ./cmd/main.go

.PHONY: run
run:
	go run ./cmd/main.go -url="localhost:9090"

.PHONY: test
test:
	go test -v ./tests
