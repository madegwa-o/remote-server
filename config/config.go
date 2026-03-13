package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config contains runtime configuration for the server.
type Config struct {

	// =========================
	// Server
	// =========================
	ServerAddr      string        `mapstructure:"SERVER_ADDR"`
	EnableTLS       bool          `mapstructure:"ENABLE_TLS"`
	TLSCertFile     string        `mapstructure:"TLS_CERT_FILE"`
	TLSKeyFile      string        `mapstructure:"TLS_KEY_FILE"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`

	// =========================
	// WebSocket
	// =========================
	ReadBufferSize  int   `mapstructure:"READ_BUFFER_SIZE"`
	WriteBufferSize int   `mapstructure:"WRITE_BUFFER_SIZE"`
	MaxMessageSize  int64 `mapstructure:"MAX_MESSAGE_SIZE"`
	MaxConnections  int   `mapstructure:"MAX_CONNECTIONS"`

	// =========================
	// Security
	// =========================
	GatewayToken string `mapstructure:"GATEWAY_TOKEN"`

	// =========================
	// Event Pipeline
	// =========================
	EventBufferSize     int `mapstructure:"EVENT_BUFFER_SIZE"`
	BroadcastBufferSize int `mapstructure:"BROADCAST_BUFFER_SIZE"`

	// =========================
	// Storage
	// =========================
	MongoURI        string `mapstructure:"MONGO_URI"`
	MongoDatabase   string `mapstructure:"MONGO_DATABASE"`
	MongoCollection string `mapstructure:"MONGO_COLLECTION"`
	StorageWorkers  int    `mapstructure:"STORAGE_WORKERS"`

	// =========================
	// Rate Limiting
	// =========================
	MaxMessagesPerSecond int `mapstructure:"MAX_MESSAGES_PER_SECOND"`
}

// Load reads configuration from environment variables and defaults.
func Load() (Config, error) {

	v := viper.New()
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// =========================
	// Defaults
	// =========================

	// Server
	v.SetDefault("SERVER_ADDR", ":8080")
	v.SetDefault("ENABLE_TLS", false)
	v.SetDefault("TLS_CERT_FILE", "")
	v.SetDefault("TLS_KEY_FILE", "")
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")

	// WebSocket
	v.SetDefault("READ_BUFFER_SIZE", 1024)
	v.SetDefault("WRITE_BUFFER_SIZE", 1024)
	v.SetDefault("MAX_MESSAGE_SIZE", 4096)
	v.SetDefault("MAX_CONNECTIONS", 10000)

	// Security
	v.SetDefault("GATEWAY_TOKEN", "dev-gateway-token")

	// Event Pipeline
	v.SetDefault("EVENT_BUFFER_SIZE", 10000)
	v.SetDefault("BROADCAST_BUFFER_SIZE", 2048)

	// Storage
	v.SetDefault("MONGO_URI", "mongodb://localhost:27017")
	v.SetDefault("MONGO_DATABASE", "telemetry")
	v.SetDefault("MONGO_COLLECTION", "vehicle_positions")
	v.SetDefault("STORAGE_WORKERS", 8)

	// Rate Limiting
	v.SetDefault("MAX_MESSAGES_PER_SECOND", 5)

	// =========================
	// Parse duration
	// =========================

	shutdownTimeout, err := time.ParseDuration(v.GetString("SHUTDOWN_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid APP_SHUTDOWN_TIMEOUT: %w", err)
	}

	v.Set("SHUTDOWN_TIMEOUT", shutdownTimeout)

	// =========================
	// Unmarshal config
	// =========================

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// =========================
	// Validation
	// =========================

	if cfg.ServerAddr == "" {
		return Config{}, fmt.Errorf("APP_SERVER_ADDR cannot be empty")
	}

	if cfg.ReadBufferSize <= 0 {
		return Config{}, fmt.Errorf("APP_READ_BUFFER_SIZE must be > 0")
	}

	if cfg.WriteBufferSize <= 0 {
		return Config{}, fmt.Errorf("APP_WRITE_BUFFER_SIZE must be > 0")
	}

	if cfg.MaxConnections <= 0 {
		return Config{}, fmt.Errorf("APP_MAX_CONNECTIONS must be > 0")
	}

	if cfg.MaxMessageSize <= 0 {
		return Config{}, fmt.Errorf("APP_MAX_MESSAGE_SIZE must be > 0")
	}

	if cfg.EventBufferSize <= 0 {
		return Config{}, fmt.Errorf("APP_EVENT_BUFFER_SIZE must be > 0")
	}

	if cfg.BroadcastBufferSize <= 0 {
		return Config{}, fmt.Errorf("APP_BROADCAST_BUFFER_SIZE must be > 0")
	}

	if cfg.StorageWorkers <= 0 {
		return Config{}, fmt.Errorf("APP_STORAGE_WORKERS must be > 0")
	}

	if cfg.MongoURI == "" {
		return Config{}, fmt.Errorf("APP_MONGO_URI cannot be empty")
	}

	if cfg.EnableTLS {
		if cfg.TLSCertFile == "" || cfg.TLSKeyFile == "" {
			return Config{}, fmt.Errorf("TLS enabled but APP_TLS_CERT_FILE or APP_TLS_KEY_FILE missing")
		}
	}

	return cfg, nil
}
