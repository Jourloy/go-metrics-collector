package storage

// Interface for work with storage
type Storage interface {
	// Update the gauge metric with the given name and value in the MemStorage struct.
	UpdateGaugeMetric(name string, value float64) error
	// Update the counter metric in the MemStorage.
	UpdateCounterMetric(name string, value int64) error
	// Return the gauge and counter maps from the MemStorage struct.
	GetValues() (map[string]float64, map[string]int64)
	// Retrieve the value of a counter from the MemStorage.
	GetCounterValue(name string) (int64, bool)
	// Return the value of a gauge by its name.
	GetGaugeValue(name string) (float64, bool)
}
