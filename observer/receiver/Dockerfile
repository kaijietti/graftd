# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

COPY main.go ./

RUN go build -o /vizor main.go