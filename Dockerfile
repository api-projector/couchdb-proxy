FROM golang:1.17.3-alpine3.13 as builder

ENV CGO_ENABLED=0

WORKDIR /go/src/couchdb-proxy
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GO111MODULE=on go build -i -v -a -installsuffix cgo -o app couchdb-proxy/cmd/server

###

FROM alpine:3.13
RUN apk add --no-cache \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 \
        -s /sbin/false \
        -G app \
        -H -h /app  \
        app

WORKDIR /app
COPY --from=builder /go/src/couchdb-proxy/app .

USER app

CMD ["./app"]