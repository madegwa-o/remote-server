# Remote Server Telemetry Platform

Scalable real-time telemetry ingestion platform in Go for vehicle data streams.

## Features

- WebSocket ingestion endpoint for edge gateways: `/ws/ingest`
- Packet validation and token authentication
- Internal event fan-out to storage and broadcast pipelines
- MongoDB persistence with geospatial schema and indexes
- Live WebSocket dashboard broadcast endpoint: `/ws/live`
- Worker-pool based storage pipeline for high throughput
- TLS-ready server configuration

## Telemetry Payload

```json
{
  "id": "17",
  "lat": -1.2921,
  "lng": 36.8219,
  "s": 42,
  "t": 1710240012
}
```

## Architecture

- **Ingestion Service**: `internal/ingestion/ws_handler.go`
- **Storage Service**: `internal/storage/mongo_store.go`
- **Broadcast Service**: `internal/broadcast/hub.go`
- **Router/Composition**: `internal/server/router.go`

Telemetry flows through an internal channel bus:

1. Gateway sends packet to `/ws/ingest`
2. Ingestion validates and pushes to `ingestCh`
3. Dispatcher fans out to:
   - `storeCh` (MongoDB write workers)
   - `liveCh` (dashboard hub broadcast)

## Local Run (Docker)

```bash
docker compose up --build
```

Services:

- App: `http://localhost:8080`
- MongoDB: `mongodb://localhost:27017`

## Environment Variables

- `APP_SERVER_ADDR` (default `:8080`)
- `APP_MONGO_URI` (default `mongodb://localhost:27017`)
- `APP_MONGO_DATABASE` (default `telemetry`)
- `APP_MONGO_COLLECTION` (default `vehicle_positions`)
- `APP_GATEWAY_TOKEN` (default `dev-gateway-token`)
- `APP_STORAGE_WORKERS` (default `8`)
- `APP_EVENT_BUFFER_SIZE` (default `10000`)
- `APP_BROADCAST_BUFFER_SIZE` (default `2048`)
- `APP_ENABLE_TLS` (default `false`)
- `APP_TLS_CERT_FILE`
- `APP_TLS_KEY_FILE`

## Quick WebSocket test

Use [`websocat`](https://github.com/vi/websocat):

Ingest:

```bash
websocat -H="Authorization: Bearer dev-gateway-token" ws://localhost:8080/ws/ingest
```

Live dashboard:

```bash
websocat ws://localhost:8080/ws/live
```
