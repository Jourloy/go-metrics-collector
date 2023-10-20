package repository

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func CreateRepository() storage.Storage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// Return the gauge and counter maps from the MemStorage struct.
func (r *MemStorage) GetValues() (map[string]float64, map[string]int64) {
	return r.gauge, r.counter
}

// Retrieve the value of a counter from the MemStorage.
func (r *MemStorage) GetCounterValue(name string) (int64, bool) {
	value, ok := r.counter[name]
	return value, ok
}

// Return the value of a gauge by its name.
func (r *MemStorage) GetGaugeValue(name string) (float64, bool) {
	value, ok := r.gauge[name]
	return value, ok
}

// Update the gauge metric with the given name and value in the MemStorage struct.
func (r *MemStorage) UpdateGaugeMetric(name string, value float64) error {
	r.gauge[name] = value

	// Prepare for difficult database
	return nil
}

// Update the counter metric in the MemStorage.
func (r *MemStorage) UpdateCounterMetric(name string, value int64) error {
	r.counter[name] += value

	// Prepare for difficult database
	return nil
}
