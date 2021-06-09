.PHONY: test prereqs post-install-test

all: test

prereqs:
	(which stringer > /dev/null) || go get golang.org/x/tools/cmd/stringer
	(which jsonenums > /dev/null) || go get github.com/campoy/jsonenums

.PHONY: unit-test
unit-test: prereqs
	cd core && go generate
	go test ./...

.PHONY: expect-test
expect-test:
	cd cmd/mexpect && go install
	cd cmd/sio && go install
	for T in `ls specs/tests/*.yaml`; do echo expecting $$T; mexpect -f $$T sio; done

test:	unit-test expect-test

install: prereqs
	go install ./...

releases:
	./tools/release.sh darwin
	./tools/release.sh windows
	./tools/release.sh linux
	./tools/release.sh linux arm

.PHONY: clean
clean:
	cd js && make clean

.PHONY: distclean
distclean: clean
	cd js && make distclean
