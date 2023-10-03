package repository

import (
	"github.com/Jourloy/go-metrics-collector/cmd/server/storage"
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

// Return values
func (r *MemStorage) ReturnValues() (map[string]float64, map[string]int64) {
	return r.gauge, r.counter
}

// Update gauge metric
func (r *MemStorage) UpdateGaugeMetric(name string, value float64) error {
	r.gauge[name] = value

	// Prepare for difficult database
	return nil
}

// Update counter metric
func (r *MemStorage) UpdateCounterMetric(name string, value int64) error {
	if r.counter[name] == 0 {
		r.counter[name] = value
	} else {
		r.counter[name] += value
	}

	// Prepare for difficult database
	return nil
}
