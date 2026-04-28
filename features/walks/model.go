package walks

import "time"

type Walk struct {
	ID                  int       `db:"id"`
	ExternalID          string    `db:"external_id"`
	StartDate           time.Time `db:"start_date"`
	EndDate             time.Time `db:"end_date"`
	DurationSeconds     float64   `db:"duration_seconds"`
	DistanceKm          float64   `db:"distance_km"`
	ElevationGainMeters float64   `db:"elevation_gain_meters"`
}

type MonthlyWalkingSpeed struct {
	Month       time.Time `db:"month"`
	AvgSpeedKmh float64   `db:"avg_speed_kmh"`
}
