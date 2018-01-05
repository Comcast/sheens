.PHONY: test prereqs

all: test

prereqs:
	@(which stringer > /dev/null) || go get golang.org/x/tools/cmd/stringer
	@(which jsonenums > /dev/null) || go get github.com/campoy/jsonenums

test: prereqs
	@cd core && go generate && go test
	@cd crew && go test
	@cd tools && go test
	@cd interpreters/goja && go test
	@cd cmd/patmatch && go test
	@cd cmd/mservice/storage/bolt && go test
	@cd cmd/mservice && go test
	@cd cmd/spectool && go test

install: prereqs
	@go install cmd/...
