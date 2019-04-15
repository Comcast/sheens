.PHONY: test prereqs post-install-test

all: test

prereqs:
	(which stringer > /dev/null) || go get golang.org/x/tools/cmd/stringer
	(which jsonenums > /dev/null) || go get github.com/campoy/jsonenums

test: prereqs
	cd core && go generate && go test
	cd crew && go test
	cd tools && go get . && go test
	cd tools/expect && go test
	cd interpreters/ecmascript && go test
	cd interpreters/noop && go test
	cd sio && go test
	for d in `find cmd -maxdepth 1 -type d  | grep /`; do (cd $$d && go test); done

install: prereqs
	for d in `find cmd -maxdepth 1 -type d  | grep /`; do (cd $$d && go install); done

post-install-test:
	for T in `ls specs/tests/*.yaml`; do echo expecting $$T; mexpect -show-err -f $$T siostd -tags=false; done


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
