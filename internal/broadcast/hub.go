package broadcast

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"remote-server/internal/models"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub broadcasts telemetry updates to connected dashboard clients.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]struct{}
	incoming   <-chan models.TelemetryEvent
	logger     zerolog.Logger
}

func NewHub(incoming <-chan models.TelemetryEvent, logger zerolog.Logger) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]struct{}),
		incoming:   incoming,
		logger:     logger.With().Str("component", "broadcast").Logger(),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAllClients()
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
			h.logger.Debug().Int("active_clients", len(h.clients)).Msg("dashboard connected")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				_ = client.conn.Close()
				h.logger.Debug().Int("active_clients", len(h.clients)).Msg("dashboard disconnected")
			}
		case event := <-h.incoming:
			h.broadcast(event)
		}
	}
}

func (h *Hub) RegisterConn(conn *websocket.Conn, queueSize int) {
	client := &Client{conn: conn, send: make(chan []byte, queueSize)}
	h.register <- client

	go h.readLoop(client)
	go h.writeLoop(client)
}

func (h *Hub) readLoop(client *Client) {
	defer func() { h.unregister <- client }()
	for {
		if _, _, err := client.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Hub) writeLoop(client *Client) {
	ticker := time.NewTicker(25 * time.Second)
	defer func() {
		ticker.Stop()
		h.unregister <- client
	}()

	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) broadcast(event models.TelemetryEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error().Err(err).Msg("marshal broadcast event")
		return
	}

	for client := range h.clients {
		select {
		case client.send <- payload:
		default:
			h.logger.Warn().Msg("slow dashboard client dropped")
			delete(h.clients, client)
			close(client.send)
			_ = client.conn.Close()
		}
	}
}

func (h *Hub) closeAllClients() {
	for client := range h.clients {
		close(client.send)
		_ = client.conn.Close()
		delete(h.clients, client)
	}
}
