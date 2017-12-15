.PHONY: test install

all:	test

test:
	cd core && go get golang.org/x/tools/cmd/stringer && go generate && go test
	cd crew && go test 
	cd tools && go test
	cd interpreters/goja && go test
	cd cmd/patmatch && go test
	cd cmd/mservice/storage/bolt && go test
	cd cmd/mservice && go test
	# cd cmd/mexpect && go test # Requires mservice executable
	cd cmd/spectool && go test

install: test
	cd cmd/patmatch && go install
	cd cmd/mservice && go install
	cd cmd/mexpect && go install
	cd cmd/spectool && go install
