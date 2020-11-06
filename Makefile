# Targets
#
.PHONY: debug
debug:	### Debug Makefile itself
	@echo $(UNAME)

.PHONY: build
build:
	go build -o wkforms ./cmd/main.go

.PHONY: run
run:
	go run ./cmd/main.go -p="9090"