FROM golang:1.18-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/faucet-svc
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/faucet-svc /go/src/faucet-svc


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/faucet-svc /usr/local/bin/faucet-svc
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["faucet-svc"]
