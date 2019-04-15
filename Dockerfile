FROM golang:latest

RUN mkdir -p $GOPATH/src/github.com/Comcast/sheens

COPY . $GOPATH/src/github.com/Comcast/sheens

WORKDIR $GOPATH/src/github.com/Comcast/sheens

RUN go get ./... && make prereqs
