FROM golang:latest

COPY . /sheens

WORKDIR /sheens

RUN make prereqs
