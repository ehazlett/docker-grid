all: deps build

deps:
	@godep restore
build:
	@godep go build .

.PHONY: build
