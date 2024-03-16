FROM golang:1.20.14

RUN mkdir -p $GOPATH/src/github.com/sheens

COPY . $GOPATH/src/github.com/sheens

WORKDIR $GOPATH/src/github.com/sheens

RUN go get ./... && make prereqs
