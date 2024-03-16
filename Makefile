.PHONY: test prereqs post-install-test

all: test

prereqs:
	(which stringer > /dev/null)  || go install golang.org/x/tools/cmd/stringer@latest
	(which jsonenums > /dev/null) || go install github.com/campoy/jsonenums@latest

test: prereqs
	cd core && go generate && go test
	cd crew && go test
	cd tools && go test
	cd tools/expect && go test
	cd interpreters/ecmascript && go test
	cd interpreters/noop && go test
	cd cmd/patmatch && go test
	cd cmd/msimple && go test
	cd cmd/mexpect && go test
	cd cmd/mcrew && go test
	cd cmd/spectool && go test
	cd cmd/sheensio && go test
	cd cmd/mdb && go test

install: prereqs
	go install cmd/...

post-install-test:
	mkdir -p tmp
	set -e; for TEST in $$(ls specs/tests/*.yaml); do echo "Running $$TEST"; mexpect -f $$TEST > tmp/$$(basename $$TEST.log) 2>&1 || (echo $$TEST failed; exit 1); done
