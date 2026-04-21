FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gophermart ./cmd/gophermart

FROM alpine:3.23

WORKDIR /app

RUN adduser -D -g '' appuser

COPY --from=builder /out/gophermart /app/gophermart
COPY migrations /app/migrations

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/gophermart"]
