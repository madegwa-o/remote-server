FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o telemetry-server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/telemetry-server ./telemetry-server
EXPOSE 9090
CMD ["./telemetry-server"]
