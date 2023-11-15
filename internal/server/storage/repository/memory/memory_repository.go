package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	StoreInterval   time.Duration
	FileStoragePath string
	IsSave          = true
	SyncSave        = false
)

type Options struct {
	StoreInterval   time.Duration
	FileStoragePath *string
	Restore         *bool
}

type MemStorage struct {
	done chan struct{}
	sync.Mutex
	gauge   map[string]float64
	counter map[string]int64
}

// CreateRepository creates a new storage repository.
//
// Returns:
// - a pointer to a storage.Storage interface.
func CreateRepository(opt Options) *MemStorage {
	StoreInterval = opt.StoreInterval

	gauge := make(map[string]float64)
	counter := make(map[string]int64)

	// Check extension and if empty add .json
	fileExt := filepath.Ext(*opt.FileStoragePath)
	if fileExt == `` {
		*opt.FileStoragePath += `.json`
	}

	FileStoragePath = *opt.FileStoragePath

	// Open file
	file, err := os.OpenFile(*opt.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		zap.L().Error(`File open error`, zap.Error(err))
	}
	defer file.Close()

	// If restore is true and file exist decode content
	if *opt.Restore && err == nil {
		var data struct {
			Gauge   *map[string]float64
			Counter *map[string]int64
		}

		err := json.NewDecoder(file).Decode(&data)
		if err != nil {
			zap.L().Error(`File decode error`, zap.Error(err))
		}

		if data.Gauge != nil {
			gauge = *data.Gauge
		}
		if data.Counter != nil {
			counter = *data.Counter
		}

		if len(gauge) > 0 || len(counter) > 0 {
			zap.L().Info(
				`MemStorage restored`,
				zap.Int(`Gauge`, len(gauge)),
				zap.Int(`Counter`, len(counter)),
			)
		}
	}

	// If StoreInterval is equal to 0, save syncronously
	if opt.StoreInterval == 0 {
		SyncSave = true
	}

	// If storage path is empty, don't save
	if *opt.FileStoragePath == `` {
		IsSave = false
	}

	return &MemStorage{
		gauge:   gauge,
		counter: counter,
		done:    make(chan struct{}),
	}
}

// StartTickers starts the tickers for the MemStorage.
func (r *MemStorage) StartTickers() {
	if SyncSave {
		return
	}

	zap.L().Info(
		`MemStorage's tickers started`,
		zap.Duration(`StoreInterval`, StoreInterval),
		zap.Bool(`Sync`, SyncSave),
	)

	saveTicker := time.NewTicker(time.Duration(StoreInterval))
	defer saveTicker.Stop()

	for {
		select {
		case <-r.done:
			return
		case <-saveTicker.C:
			if !SyncSave {
				r.SaveMetricsOnDisk()
			}
		}
	}
}

// SaveMetricsOnDisk saves the metrics in memory to a file on disk.
func (r *MemStorage) SaveMetricsOnDisk() {
	if !IsSave {
		return
	}

	if _, err := os.Stat(FileStoragePath); os.IsNotExist(err) {
		zap.L().Warn(`File doesn't exist`)
	} else {
		if err := os.Truncate(FileStoragePath, 0); err != nil {
			zap.L().Error(`File truncate error`, zap.Error(err))
			return
		}
	}

	file, err := os.OpenFile(FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}
	defer file.Close()

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	data := make(map[string]any)
	data[`gauge`] = r.gauge
	data[`counter`] = r.counter

	if err := json.NewEncoder(file).Encode(data); err != nil {
		zap.L().Error(`File encode error`, zap.Error(err))
		return
	}

	zap.L().Debug(`Metrics saved on disk`)
}

// GetValues returns the gauge and counter maps of the MemStorage.
//
// Returns:
// - map[string]float64
// - map[string]int64.
func (r *MemStorage) GetValues() (map[string]float64, map[string]int64) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

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
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

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
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

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
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.gauge[name] = value

	// Save metrics on disk if SyncSave is true
	if SyncSave {
		go r.SaveMetricsOnDisk()
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
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.counter[name] += value

	// Save metrics on disk if SyncSave is true
	if SyncSave {
		go r.SaveMetricsOnDisk()
	}

	return r.counter[name]
}
