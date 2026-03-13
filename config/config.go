package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config contains runtime configuration for the server.
type Config struct {
	ServerAddr          string
	ReadBufferSize      int
	WriteBufferSize     int
	MongoURI            string
	MongoDatabase       string
	MongoCollection     string
	GatewayToken        string
	EventBufferSize     int
	StorageWorkers      int
	BroadcastBufferSize int
	ShutdownTimeout     time.Duration
	EnableTLS           bool
	TLSCertFile         string
	TLSKeyFile          string
}

// Load reads configuration from env vars and defaults.
func Load() (Config, error) {
	v := viper.New()
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	v.SetDefault("SERVER_ADDR", ":8080")
	v.SetDefault("READ_BUFFER_SIZE", 1024)
	v.SetDefault("WRITE_BUFFER_SIZE", 1024)
	v.SetDefault("MONGO_URI", "mongodb://localhost:27017")
	v.SetDefault("MONGO_DATABASE", "telemetry")
	v.SetDefault("MONGO_COLLECTION", "vehicle_positions")
	v.SetDefault("GATEWAY_TOKEN", "dev-gateway-token")
	v.SetDefault("EVENT_BUFFER_SIZE", 10000)
	v.SetDefault("STORAGE_WORKERS", 8)
	v.SetDefault("BROADCAST_BUFFER_SIZE", 2048)
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")
	v.SetDefault("ENABLE_TLS", false)
	v.SetDefault("TLS_CERT_FILE", "")
	v.SetDefault("TLS_KEY_FILE", "")

	shutdownTimeout, err := time.ParseDuration(v.GetString("SHUTDOWN_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid APP_SHUTDOWN_TIMEOUT: %w", err)
	}

	cfg := Config{
		ServerAddr:          v.GetString("SERVER_ADDR"),
		ReadBufferSize:      v.GetInt("READ_BUFFER_SIZE"),
		WriteBufferSize:     v.GetInt("WRITE_BUFFER_SIZE"),
		MongoURI:            v.GetString("MONGO_URI"),
		MongoDatabase:       v.GetString("MONGO_DATABASE"),
		MongoCollection:     v.GetString("MONGO_COLLECTION"),
		GatewayToken:        v.GetString("GATEWAY_TOKEN"),
		EventBufferSize:     v.GetInt("EVENT_BUFFER_SIZE"),
		StorageWorkers:      v.GetInt("STORAGE_WORKERS"),
		BroadcastBufferSize: v.GetInt("BROADCAST_BUFFER_SIZE"),
		ShutdownTimeout:     shutdownTimeout,
		EnableTLS:           v.GetBool("ENABLE_TLS"),
		TLSCertFile:         v.GetString("TLS_CERT_FILE"),
		TLSKeyFile:          v.GetString("TLS_KEY_FILE"),
	}

	if cfg.StorageWorkers <= 0 {
		return Config{}, fmt.Errorf("APP_STORAGE_WORKERS must be > 0")
	}
	if cfg.EventBufferSize <= 0 {
		return Config{}, fmt.Errorf("APP_EVENT_BUFFER_SIZE must be > 0")
	}
	if cfg.BroadcastBufferSize <= 0 {
		return Config{}, fmt.Errorf("APP_BROADCAST_BUFFER_SIZE must be > 0")
	}

	if cfg.EnableTLS && (cfg.TLSCertFile == "" || cfg.TLSKeyFile == "") {
		return Config{}, fmt.Errorf("TLS enabled but APP_TLS_CERT_FILE or APP_TLS_KEY_FILE missing")
	}

	return cfg, nil
}
