package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"remote-server/internal/models"
)

type geoPoint struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

type telemetryDocument struct {
	VehicleID string   `bson:"vehicleId"`
	Timestamp int64    `bson:"timestamp"`
	Location  geoPoint `bson:"location"`
	Speed     float64  `bson:"speed"`
}

// MongoStore persists telemetry events in MongoDB.
type MongoStore struct {
	client     *mongo.Client
	collection *mongo.Collection
	logger     zerolog.Logger
}

func NewMongoStore(ctx context.Context, uri, dbName, collectionName string, logger zerolog.Logger) (*MongoStore, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	col := client.Database(dbName).Collection(collectionName)
	if err := ensureIndexes(ctx, col); err != nil {
		return nil, fmt.Errorf("ensure indexes: %w", err)
	}

	return &MongoStore{client: client, collection: col, logger: logger.With().Str("component", "storage").Logger()}, nil
}

func ensureIndexes(ctx context.Context, collection *mongo.Collection) error {
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "vehicleId", Value: 1}, {Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "location", Value: "2dsphere"}}},
	}
	_, err := collection.Indexes().CreateMany(ctx, models)
	return err
}

func (m *MongoStore) Store(ctx context.Context, event models.TelemetryEvent) error {
	doc := telemetryDocument{
		VehicleID: event.VehicleID,
		Timestamp: event.Timestamp,
		Location: geoPoint{
			Type:        "Point",
			Coordinates: []float64{event.Lng, event.Lat},
		},
		Speed: event.Speed,
	}
	_, err := m.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert telemetry: %w", err)
	}
	return nil
}

func (m *MongoStore) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
