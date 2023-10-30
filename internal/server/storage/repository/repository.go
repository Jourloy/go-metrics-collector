package repository

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"go.uber.org/zap"
)

var (
	StoreInterval   = flag.Int(`i`, 300, "Store interval in seconds")
	FileStoragePath = flag.String(`f`, `/tmp/metrics-db.json`, "File storage path")
	Restore         = flag.Bool(`r`, true, "Restore from file")
	isSave          = true
	syncSave        = false
)

func init() {
	if env, exist := os.LookupEnv(`STORE_INTERVAL`); exist {
		if i, err := strconv.Atoi(env); err == nil {
			StoreInterval = &i
		}
	}

	if env, exist := os.LookupEnv(`FILE_STORAGE_PATH`); exist {
		FileStoragePath = &env
	}

	if env, exist := os.LookupEnv(`RESTORE`); exist {
		if b, err := strconv.ParseBool(env); err == nil {
			Restore = &b
		}
	}
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
	done    chan struct{}
}

// CreateRepository creates a new storage repository.
//
// Reutrns:
// - a pointer to a storage.Storage interface.
func CreateRepository() storage.Storage {
	gauge := make(map[string]float64)
	counter := make(map[string]int64)

	// Open file
	file, err := os.OpenFile(*FileStoragePath, os.O_RDONLY, 0666)
	if err != nil {
		zap.L().Error(err.Error())
	}
	defer file.Close()

	// If restore is true and file exist decode content
	if *Restore && err == nil {
		var data struct {
			Gauge   map[string]float64
			Counter map[string]int64
		}

		if err := json.NewDecoder(file).Decode(&data); err != nil {
			zap.L().Error(err.Error())
		}
		if data.Gauge != nil {
			gauge = data.Gauge
		}
		if data.Counter != nil {
			counter = data.Counter
		}
	}

	// If StoreInterval is equal to 0, save syncronously
	if *StoreInterval == 0 {
		syncSave = true
	}

	// If storage path is empty, don't save
	if *FileStoragePath == `` {
		isSave = false
	}

	return &MemStorage{
		gauge:   gauge,
		counter: counter,
		done:    make(chan struct{}),
	}
}

// StartTickers starts the tickers for the MemStorage.
//
// No parameters.
// No return values.
func (r *MemStorage) StartTickers() {
	saveTicker := time.NewTicker(time.Duration(*StoreInterval) * time.Second)

	for {
		select {
		case <-r.done:
			return
		case <-saveTicker.C:
			if !syncSave {
				r.SaveMetricsOnDisk()
			}
		}
	}
}

// StopTickers stops the tickers in the MemStorage.
//
// No parameters.
// No return types.
func (r *MemStorage) StopTickers() {
	close(r.done)
}

// SaveMetricsOnDisk saves the metrics in memory to a file on disk.
//
// No parameters.
// No return types.
func (r *MemStorage) SaveMetricsOnDisk() {
	if !isSave {
		return
	}

	if _, err := os.Stat(*FileStoragePath); os.IsNotExist(err) {
		zap.L().Warn(`File doesn't exist`)
	} else {
		if err := os.Truncate(*FileStoragePath, 0); err != nil {
			zap.L().Error(err.Error())
			return
		}
	}

	file, err := os.OpenFile(*FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}
	defer file.Close()

	data := make(map[string]any)
	data["gauge"] = r.gauge
	data["counter"] = r.counter

	if err := json.NewEncoder(file).Encode(data); err != nil {
		zap.L().Error(err.Error())
		return
	}

	zap.L().Debug(`Metrics saved on disk`)
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
	if syncSave {
		r.SaveMetricsOnDisk()
	}
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
	if syncSave {
		r.SaveMetricsOnDisk()
	}
	return r.counter[name]
}
