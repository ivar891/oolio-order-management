# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    go build -ldflags='-w -s' \
    -o /server ./cmd/server/

# Stage 2: Runtime — minimal distroless image
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/migrations /migrations
COPY --from=builder /server /server

WORKDIR /

EXPOSE 8080 8081

USER nonroot:nonroot

ENTRYPOINT ["/server"]
