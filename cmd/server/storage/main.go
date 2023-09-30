package storage

// Interface for work with storage
type Storage interface {
	UpdateGaugeMetric(name string, value float64) error
	UpdateCounterMetric(name string, value int64) error
	ReturnValues() (map[string]float64, map[string]int64)
}
