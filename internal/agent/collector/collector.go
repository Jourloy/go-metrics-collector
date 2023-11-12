package collector

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var (
	ServerAddress  = flag.String("a", `localhost:8080`, "Host of the server")
	ReportInterval = flag.Int("r", 5, "Report Interval")
	PollInterval   = flag.Int("p", 2, "Poll Interval")
)

type Collector struct {
	done    chan struct{}
	gauge   map[string]float64
	counter map[string]int64
}

type Metric struct {
	ID    string   `json:"id"`              // Name of metric
	MType string   `json:"type"`            // Gauge or Counter
	Delta *int64   `json:"delta,omitempty"` // Value if metric is a counter
	Value *float64 `json:"value,omitempty"` // Value if metric is a gauge
}

// envParse initializes the ServerAddress, PollInterval, and ReportInterval
// variables by checking for corresponding environment variables.
func envParse() {
	if hostENV, exist := os.LookupEnv(`ADDRESS`); exist {
		ServerAddress = &hostENV
	}

	if pollENV, exist := os.LookupEnv(`POLL_INTERVAL`); exist {
		if i, err := strconv.Atoi(pollENV); err == nil {
			PollInterval = &i
		}
	}

	if reportENV, exist := os.LookupEnv(`REPORT_INTERVAL`); exist {
		if i, err := strconv.Atoi(reportENV); err == nil {
			ReportInterval = &i
		}
	}

	zap.L().Debug(`Collector initialized`)
}

// CreateCollector creates a new instance of the Collector struct.
//
// Returns:
// - a pointer to a Collector.
func CreateCollector() *Collector {
	// Parse environment variables
	envParse()

	return &Collector{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
		done:    make(chan struct{}),
	}
}

// StartTickers starts the tickers for collecting and sending metrics in the Collector struct.
func (c *Collector) StartTickers() {
	// Start tickers
	collectTicker := time.NewTicker(time.Duration(*PollInterval) * time.Second)
	defer collectTicker.Stop()

	sendTicker := time.NewTicker(time.Duration(*ReportInterval) * time.Second)
	defer sendTicker.Stop()

	zap.L().Info(`Collector's tickers started`)

	for {
		select {
		case <-c.done:
			return
		case <-collectTicker.C:
			c.collectMetric()
		case <-sendTicker.C:
			c.sendMetrics()
		}
	}
}

// CloseChannel close channel and as a result stops the tickers of the Collector.
func (c *Collector) CloseChannel() {
	zap.L().Info(`Collector's tickers stopped`)
	close(c.done)
}

// collectMetric collects various metrics and stores them in the gauge and counter maps.
func (c *Collector) collectMetric() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.gauge[`Alloc`] = float64(memStats.Alloc)
	c.gauge[`BuckHashSys`] = float64(memStats.BuckHashSys)
	c.gauge[`Frees`] = float64(memStats.Frees)
	c.gauge[`GCCPUFraction`] = float64(memStats.GCCPUFraction)
	c.gauge[`GCSys`] = float64(memStats.GCSys)
	c.gauge[`HeapAlloc`] = float64(memStats.HeapAlloc)
	c.gauge[`HeapIdle`] = float64(memStats.HeapIdle)
	c.gauge[`HeapInuse`] = float64(memStats.HeapInuse)
	c.gauge[`HeapReleased`] = float64(memStats.HeapReleased)
	c.gauge[`HeapObjects`] = float64(memStats.HeapObjects)
	c.gauge[`HeapSys`] = float64(memStats.HeapSys)
	c.gauge[`LastGC`] = float64(memStats.LastGC)
	c.gauge[`Lookups`] = float64(memStats.Lookups)
	c.gauge[`MCacheInuse`] = float64(memStats.MCacheInuse)
	c.gauge[`MCacheSys`] = float64(memStats.MCacheSys)
	c.gauge[`MSpanInuse`] = float64(memStats.MSpanInuse)
	c.gauge[`MSpanSys`] = float64(memStats.MSpanSys)
	c.gauge[`Mallocs`] = float64(memStats.Mallocs)
	c.gauge[`NextGC`] = float64(memStats.NextGC)
	c.gauge[`NumForcedGC`] = float64(memStats.NumForcedGC)
	c.gauge[`NumGC`] = float64(memStats.NumGC)
	c.gauge[`OtherSys`] = float64(memStats.OtherSys)
	c.gauge[`PauseTotalNs`] = float64(memStats.PauseTotalNs)
	c.gauge[`StackInuse`] = float64(memStats.StackInuse)
	c.gauge[`StackSys`] = float64(memStats.StackSys)
	c.gauge[`Sys`] = float64(memStats.Sys)
	c.gauge[`TotalAlloc`] = float64(memStats.TotalAlloc)
	c.gauge[`RandomValue`] = rand.Float64()

	c.counter[`PollCount`]++

	zap.L().Debug(`Metrics collected`)
}

type Statuses struct {
	Success  int
	Internal int
	Fail     int
}

// sendMetrics sends the metrics to the server.
func (c *Collector) sendMetrics() {
	statuses := Statuses{}
	metrics := []Metric{}

	for name, value := range c.gauge {
		metrics = append(metrics, Metric{
			ID:    name,
			MType: `gauge`,
			Value: &value,
		})
	}

	for name, value := range c.counter {
		metrics = append(metrics, Metric{
			ID:    name,
			MType: `counter`,
			Delta: &value,
		})
	}

	status, err := c.sendPOST(metrics)
	if err == nil && status != nil {
		if *status == 200 {
			statuses.Success++
		} else if *status == 500 {
			statuses.Internal++
		} else {
			statuses.Fail++
		}
	}

	// Reset poll count
	c.counter[`PollCount`] = 0

	// Log sent metrics
	zap.L().Debug(
		`Metrics sent to the server`,
		zap.String(`Success`, fmt.Sprintf(`%d`, statuses.Success)),
		zap.String(`Internal`, fmt.Sprintf(`%d`, statuses.Internal)),
		zap.String(`Fail`, fmt.Sprintf(`%d`, statuses.Fail)),
	)
}

// sendPOST sends a POST request to the server with the given metric.
//
// Parameters:
// - metric: the metric to be sent
func (c *Collector) sendPOST(metrics []Metric) (*int, error) {
	b, _ := json.Marshal(metrics)

	var gz bytes.Buffer

	w := gzip.NewWriter(&gz)
	w.Write(b)
	w.Close()

	req, err := http.NewRequest(http.MethodPost, `http://`+*ServerAddress+`/updates/`, &gz)
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}

	req.Header.Set(`Content-Encoding`, `gzip`)
	req.Header.Set(`Accept-Encoding`, `gzip`)
	req.Header.Set(`Content-Type`, `application/json`)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}
	defer res.Body.Close()

	return &res.StatusCode, nil
}
