package repository

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// CreateRepository creates a new storage repository.
//
// Reutrns:
// - a pointer to a storage.Storage interface.
func CreateRepository() storage.Storage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// GetValues returns the gauge and counter maps of the MemStorage.
//
// No parameters.
// Returns:
// - map[string]float64
// - map[string]int64.
func (r *MemStorage) GetValues() (map[string]float64, map[string]int64) {
	return r.gauge, r.counter
}

// GetCounterValue retrieves the value of a counter by its name from the MemStorage.
//
// Parameters:
// - name: the name of the counter.
//
// Returns:
// - int64: the value of the counter.
// - bool: true if the counter exists, false otherwise.
func (r *MemStorage) GetCounterValue(name string) (int64, bool) {
	value, ok := r.counter[name]
	return value, ok
}

// GetGaugeValue retrieves the value of a gauge by its name from the MemStorage.
//
// Parameters:
// - name: a string representing the name of the gauge.
//
// Returns:
// - value: a float64 representing the value of the gauge.
// - ok: a boolean indicating whether the gauge was found.
func (r *MemStorage) GetGaugeValue(name string) (float64, bool) {
	value, ok := r.gauge[name]
	return value, ok
}

// UpdateGaugeMetric updates the gauge metric with the given name and value in the MemStorage.
//
// Parameters:
// - name: the name of the gauge metric (string)
// - value: the value of the gauge metric (float64)
//
// Returns:
// - the updated value of the gauge metric (float64).
func (r *MemStorage) UpdateGaugeMetric(name string, value float64) float64 {
	r.gauge[name] = value
	return r.gauge[name]
}

// UpdateCounterMetric updates the counter metric with the given name by adding the value to it.
//
// Parameters:
// - name: the name of the counter metric (string)
// - value: the value to be added to the counter metric (int64)
//
// Returns:
// - the updated value of the counter metric (int64)
func (r *MemStorage) UpdateCounterMetric(name string, value int64) int64 {
	r.counter[name] += value
	return r.counter[name]
}
