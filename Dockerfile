# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /graftd

COPY . ./

RUN go env -w GOPROXY="https://goproxy.io,direct"

RUN go build -mod vendor -o /raftnode