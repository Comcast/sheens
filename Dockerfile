FROM golang:latest

RUN mkdir -p $GOPATH/src/github.com/sheens

COPY . $GOPATH/src/github.com/sheens

WORKDIR $GOPATH/src/github.com/sheens

RUN go get ./... && make prereqs
