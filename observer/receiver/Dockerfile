# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /work

COPY go.mod ./
COPY go.sum ./
RUN go env -w GOPROXY="https://goproxy.io,direct"
RUN go mod download

COPY . ./

RUN go build -o /vizor