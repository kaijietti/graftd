# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go env -w GOPROXY="https://goproxy.io,direct"
RUN go mod download

COPY *.go ./

RUN go build -o raftexample

CMD [ "./raftexample -id $(hostname) ~/$(hostname) " ]
