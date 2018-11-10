FROM golang:1.11-alpine AS builder

WORKDIR /app
ENV GO111MODUlE=on CGO_ENABLED=0 GOOS=linux
RUN addgroup -S app && adduser -S -G app app

RUN apk add --no-cache git upx ca-certificates

COPY go.mod /app
RUN go mod download && go mod verify

COPY . /app
RUN go build -ldflags="-s -w" -o frontend ./cmd/frontend && upx --lzma frontend


FROM scratch
EXPOSE "8080/tcp"
ENTRYPOINT ["/frontend"]
CMD []
COPY --from=builder /etc/passwd /etc/passwd
USER app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/http/templates /http/templates
COPY --from=builder /app/frontend /frontend