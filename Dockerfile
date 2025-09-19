# Stage 1: build
FROM golang:1.24-alpine AS builder
WORKDIR /app

# tools untuk go mod + build
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cmsapp ./cmd/main.go

# Stage 2: runtime
FROM alpine:3.18
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/cmsapp .
# don't copy .env into image; we use docker-compose env_file
EXPOSE 3000
ENTRYPOINT ["./cmsapp"]
