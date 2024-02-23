export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: test prereqs post-install-test

all: test

prereqs:
	@(which stringer > /dev/null) || go install golang.org/x/tools/cmd/stringer
	@(which jsonenums > /dev/null) || go install github.com/campoy/jsonenums

test: prereqs
	go test -v ./...

install: prereqs
	go install cmd/...

post-install-test:
	mkdir -p tmp
	set -e; for TEST in $$(ls specs/tests/*.yaml); do echo "Running $$TEST"; mexpect -f $$TEST > tmp/$$(basename $$TEST.log) 2>&1 || (echo $$TEST failed; exit 1); done
