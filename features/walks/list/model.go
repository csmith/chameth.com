package list

type WalkEntry struct {
	Date             string
	DistanceBarWidth float64
	DistanceKm       float64
	ElevationGainM   float64
	Duration         string
}

type Data struct {
	Walks []WalkEntry
}
