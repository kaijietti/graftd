# syntax=docker/dockerfile:1

FROM ubuntu:18.04

ENV GOLANG_VERSION 1.17.8

# Install wget
RUN apt update
RUN apt install -y build-essential wget
RUN apt install -y iproute2
# Install Go
RUN wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz
RUN rm -f go${GOLANG_VERSION}.linux-amd64.tar.gz

ENV PATH "$PATH:/usr/local/go/bin"

WORKDIR /graftd

COPY . ./

RUN go env -w GOPROXY="https://goproxy.io,direct"

RUN go build -mod vendor -o /raftnode

RUN apt install -y kernel-modules-extra

#FROM golang:1.17-alpine
#
#RUN apk add bash
#RUN apk add iproute2
#
#WORKDIR /graftd
#
#COPY . ./
#
#RUN go env -w GOPROXY="https://goproxy.io,direct"
#
#RUN go build -mod vendor -o /raftnode