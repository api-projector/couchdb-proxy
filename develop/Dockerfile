FROM golang:1.15.2-alpine3.12 as build-env

ENV CGO_ENABLED=0

WORKDIR /go/src/couchdb-proxy
COPY go.mod go.sum ./
RUN go mod download
RUN go get github.com/go-delve/delve/cmd/dlv

COPY . .

RUN GOOS=linux GO111MODULE=on go build -gcflags="all=-N -l" -o /app couchdb-proxy/cmd/server

###

FROM debian:buster

EXPOSE 8000 40000

WORKDIR /
COPY --from=build-env /go/bin/dlv /
COPY --from=build-env /app /

CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/app"]