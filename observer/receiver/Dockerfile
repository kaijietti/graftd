# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /work

COPY . ./

RUN go env -w GOPROXY="https://goproxy.io,direct"

RUN go build -o /vizor