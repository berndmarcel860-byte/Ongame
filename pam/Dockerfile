FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /pam ./cmd/server

FROM gcr.io/distroless/static-debian12

COPY --from=builder /pam /pam
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

ENV MIGRATIONS_DIR=/migrations

ENTRYPOINT ["/pam"]
