package ingestion

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"remote-server/internal/models"
)

// WSHandler handles telemetry ingestion over WebSocket.
type WSHandler struct {
	upgrader websocket.Upgrader
	token    string
	events   chan<- models.TelemetryEvent
	logger   zerolog.Logger
}

func NewWSHandler(readBufferSize, writeBufferSize int, token string, events chan<- models.TelemetryEvent, logger zerolog.Logger) *WSHandler {
	return &WSHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		token:  token,
		events: events,
		logger: logger.With().Str("component", "ingestion").Logger(),
	}
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed upgrading ingestion websocket")
		return
	}
	defer conn.Close()

	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	h.logger.Info().Str("remote", r.RemoteAddr).Msg("gateway connected")
	defer h.logger.Info().Str("remote", r.RemoteAddr).Msg("gateway disconnected")

	for {
		var packet models.TelemetryPacket
		if err := conn.ReadJSON(&packet); err != nil {
			h.logger.Debug().Err(err).Msg("gateway read closed")
			return
		}

		if err := packet.Validate(); err != nil {
			h.logger.Warn().Err(err).Str("vehicle", packet.ID).Msg("invalid telemetry packet")
			_ = conn.WriteJSON(map[string]string{"error": err.Error()})
			continue
		}

		event := packet.ToEvent()
		select {
		case h.events <- event:
		default:
			h.logger.Error().Str("vehicle", event.VehicleID).Msg("event bus overloaded")
			_ = conn.WriteJSON(map[string]string{"error": "server overloaded"})
		}
	}
}

func (h *WSHandler) authorized(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ") == h.token
	}
	return r.URL.Query().Get("token") == h.token
}
