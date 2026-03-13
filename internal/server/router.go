package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"remote-server/config"
	"remote-server/internal/broadcast"
	"remote-server/internal/models"
	"remote-server/internal/storage"
)

// App wires all telemetry services together.
type App struct {
	cfg      config.Config
	logger   zerolog.Logger
	server   *http.Server
	store    *storage.MongoStore
	ingestCh chan models.TelemetryEvent
	storeCh  chan models.TelemetryEvent
	liveCh   chan models.TelemetryEvent
	hub      *broadcast.Hub
}

func New(ctx context.Context, cfg config.Config, logger zerolog.Logger) (*App, error) {
	store, err := storage.NewMongoStore(ctx, cfg.MongoURI, cfg.MongoDatabase, cfg.MongoCollection, logger)
	if err != nil {
		return nil, fmt.Errorf("init store: %w", err)
	}

	ingestCh := make(chan models.TelemetryEvent, cfg.EventBufferSize)
	storeCh := make(chan models.TelemetryEvent, cfg.EventBufferSize)
	liveCh := make(chan models.TelemetryEvent, cfg.EventBufferSize)

	hub := broadcast.NewHub(liveCh, logger)

	mux := http.NewServeMux()

	// -------------------------------
	// WebSocket for Edge Gateway ingestion
	// -------------------------------
	mux.HandleFunc("/ws/ingest", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  cfg.ReadBufferSize,
			WriteBufferSize: cfg.WriteBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true }, // allow all origins
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error().Err(err).Msg("failed websocket upgrade")
			return
		}
		logger.Info().Msg("new edge gateway connected")

		// Launch read loop in a goroutine
		go func() {
			defer conn.Close()
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					logger.Warn().Err(err).Msg("read failed, closing connection")
					break
				}

				var event models.TelemetryEvent
				if err := json.Unmarshal(msg, &event); err != nil {
					logger.Error().Err(err).Msgf("invalid telemetry JSON: %s", string(msg))
					continue
				}
				event.Timestamp = time.Now().UnixMilli()

				// Dispatch to ingest channel
				select {
				case ingestCh <- event:
				case <-r.Context().Done():
					return
				}
			}
		}()
	})

	// -------------------------------
	// WebSocket for live dashboards
	// -------------------------------
	mux.HandleFunc("/ws/live", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  cfg.ReadBufferSize,
			WriteBufferSize: cfg.WriteBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn().Err(err).Msg("failed upgrading live websocket")
			return
		}
		hub.RegisterConn(conn, cfg.BroadcastBufferSize)
	})

	// -------------------------------
	// Health endpoint
	// -------------------------------
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{Addr: cfg.ServerAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	return &App{
		cfg:      cfg,
		logger:   logger,
		server:   srv,
		store:    store,
		ingestCh: ingestCh,
		storeCh:  storeCh,
		liveCh:   liveCh,
		hub:      hub,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.dispatch(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.hub.Run(ctx)
	}()

	for i := 0; i < a.cfg.StorageWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			a.storageWorker(ctx, workerID)
		}(i)
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info().Str("addr", a.cfg.ServerAddr).Msg("server listening")
		if a.cfg.EnableTLS {
			errCh <- a.server.ListenAndServeTLS(a.cfg.TLSCertFile, a.cfg.TLSKeyFile)
			return
		}
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer shutdownCancel()
	_ = a.server.Shutdown(shutdownCtx)
	cancel()
	wg.Wait()
	return a.store.Close(shutdownCtx)
}

// dispatch routes events from ingestCh to storeCh and liveCh
func (a *App) dispatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-a.ingestCh:
			select {
			case a.storeCh <- event:
			case <-ctx.Done():
				return
			}
			select {
			case a.liveCh <- event:
			case <-ctx.Done():
				return
			}
		}
	}
}

// storageWorker writes events to MongoDB
func (a *App) storageWorker(ctx context.Context, workerID int) {
	logger := a.logger.With().Str("component", "storage_worker").Int("worker", workerID).Logger()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-a.storeCh:
			if err := a.store.Store(ctx, event); err != nil {
				logger.Error().Err(err).Str("vehicle", event.VehicleID).Msg("failed to persist telemetry")
			}
		}
	}
}
