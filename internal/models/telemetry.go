package models

import (
	"errors"
	"fmt"
)

// TelemetryPacket is the payload sent by edge gateways.
type TelemetryPacket struct {
	ID  string  `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	S   float64 `json:"s"`
	T   int64   `json:"t"`
}

// TelemetryEvent is the canonical internal event used for storage and broadcast.
type TelemetryEvent struct {
	VehicleID string  `json:"vehicleId" bson:"vehicleId"`
	Timestamp int64   `json:"timestamp" bson:"timestamp"`
	Lat       float64 `json:"lat" bson:"lat"`
	Lng       float64 `json:"lng" bson:"lng"`
	Speed     float64 `json:"speed" bson:"speed"`
}

// Validate ensures telemetry packet values are within acceptable ranges.
func (p TelemetryPacket) Validate() error {
	if p.ID == "" {
		return errors.New("id is required")
	}
	if p.Lat < -90 || p.Lat > 90 {
		return fmt.Errorf("lat out of range: %v", p.Lat)
	}
	if p.Lng < -180 || p.Lng > 180 {
		return fmt.Errorf("lng out of range: %v", p.Lng)
	}
	if p.S < 0 {
		return fmt.Errorf("speed must be non-negative: %v", p.S)
	}
	if p.T <= 0 {
		return fmt.Errorf("timestamp must be > 0: %d", p.T)
	}
	return nil
}

// ToEvent converts an incoming packet to the internal event shape.
func (p TelemetryPacket) ToEvent() TelemetryEvent {
	return TelemetryEvent{
		VehicleID: p.ID,
		Timestamp: p.T,
		Lat:       p.Lat,
		Lng:       p.Lng,
		Speed:     p.S,
	}
}
