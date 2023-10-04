package storage

// Interface for work with storage
type Storage interface {
	UpdateGaugeMetric(name string, value float64) error
	UpdateCounterMetric(name string, value int64) error
	GetValues() (map[string]float64, map[string]int64)
	GetCounterValue(name string) int64
	GetGaugeValue(name string) float64
}
