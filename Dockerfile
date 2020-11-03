##
# Builder
##
FROM golang:1.15 AS builder

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-sentry-cli

##
# Main image
##
FROM scratch

COPY --from=builder /app/go-sentry-cli /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/go-sentry-cli"]

CMD ["--help"]