FROM golang:1.22 AS exporter

ENV GOBIN=/go/bin
ENV GOPATH=/go
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /exporter
COPY *.go go.sum go.mod ./
RUN go build -o /da-exporter .

FROM debian:buster-slim

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates
RUN useradd -ms /bin/bash exporter && chown -R exporter /usr

EXPOSE 9300

COPY --from=exporter da-exporter /usr/bin/da-exporter

USER exporter