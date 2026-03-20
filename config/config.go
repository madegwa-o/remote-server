package config

import (
	"time"
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
	return Config{

		// =========================
		// Server
		// =========================
		ServerAddr:      ":8080",
		EnableTLS:       false,
		TLSCertFile:     "",
		TLSKeyFile:      "",
		ShutdownTimeout: 10 * time.Second,

		// =========================
		// WebSocket
		// =========================
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		MaxMessageSize:  4096,
		MaxConnections:  10000,

		// =========================
		// Security
		// =========================
		GatewayToken: "dev-gateway-token",

		// =========================
		// Event Pipeline
		// =========================
		EventBufferSize:     10000,
		BroadcastBufferSize: 2048,

		// =========================
		// Storage
		// =========================
		MongoURI:        "mongodb+srv://dev:S5D3QvmUKatE4NvL@ossycluster.gt2ff.mongodb.net/antitest?appName=OssyCluster",
		MongoDatabase:   "telemetry",
		MongoCollection: "vehicle_positions",
		StorageWorkers:  8,

		// =========================
		// Rate Limiting
		// =========================
		MaxMessagesPerSecond: 5,
	}, nil
}
