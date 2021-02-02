ARG VERSION=development
FROM golang:1.15-buster as builder
ENV GO111MODULE=on
ENV CGO_ENABLED=0
WORKDIR /usr/src/
COPY . /usr/src
RUN go build -v -ldflags="-X 'github.com/cloudposse/turf/cmd.Version=${VERSION}'" -o "bin/turf" *.go 

FROM alpine:3.13
RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/src/bin/* /usr/bin/
ENV PATH $PATH:/usr/bin
ENTRYPOINT ["turf"]