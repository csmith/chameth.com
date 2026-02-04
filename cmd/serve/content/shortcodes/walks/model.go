package walks

type WalkEntry struct {
	Date             string
	DistanceBarWidth float64 // Percentage (0-100) for progress bar
	DistanceKm       float64
	ElevationGainM   float64
	Duration         string
}

type Data struct {
	Walks []WalkEntry
}
