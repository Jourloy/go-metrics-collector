// Package collector collect OS metrics and send they to the server
//
// Get collector agent: `agent := collector.CreateCollector()`
//
// Start collector: agent.StartTickers()
package collector

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

var (
	ServerAddress  = flag.String("a", `localhost:8080`, "Host of the server")
	ReportInterval = flag.Int("r", 5, "Report Interval")
	PollInterval   = flag.Int("p", 2, "Poll Interval")
	Key            = flag.String(`k`, ``, `Key for hash`)
	RateLimit      = flag.Int(`i`, 0, `Rate limit. 0 - no limit`)
)

type Collector struct {
	done chan struct{}
	sync.Mutex
	gauge   map[string]float64
	counter map[string]int64
}

type Metric struct {
	ID    string   `json:"id"`              // Name of metric
	MType string   `json:"type"`            // Gauge or Counter
	Delta *int64   `json:"delta,omitempty"` // Value if metric is a counter
	Value *float64 `json:"value,omitempty"` // Value if metric is a gauge
}

type AgentConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
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

	if keyENV, exist := os.LookupEnv(`KEY`); exist {
		Key = &keyENV
	}

	if rateENV, exist := os.LookupEnv(`RATE_LIMIT`); exist {
		if i, err := strconv.Atoi(rateENV); err == nil {
			ReportInterval = &i
		}
	}

	if file, err := os.Open(`./agent.config.json`); err == nil {
		defer file.Close()

		b, _ := io.ReadAll(file)
		var config AgentConfig
		json.Unmarshal(b, &config)

		ServerAddress = &config.Address

		if i, err := strconv.Atoi(config.PollInterval); err == nil {
			PollInterval = &i
		}

		if i, err := strconv.Atoi(config.ReportInterval); err == nil {
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
			c.collectPsutilMetric()
		case <-sendTicker.C:
			go c.sendMetrics()
		}
	}
}

// CloseChannel close channel and as a result stops the tickers of the Collector.
func (c *Collector) CloseChannel() {
	zap.L().Info(`Collector's tickers stopped`)
	close(c.done)
}

// collectMetric collects various metrics and stores them in the gauge and counter maps.
//
// For mentor: This is already gorutine, looks like worker, so I don't change code below
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

func (c *Collector) collectPsutilMetric() {
	v, err := mem.VirtualMemory()
	if err != nil {
		zap.L().Error(`Cannot get virtual memory`, zap.Error(err))
		return
	}

	cp, err := cpu.Times(true)
	if err != nil {
		zap.L().Error(`Cannot get cpu info`, zap.Error(err))
		return
	}

	c.gauge[`TotalMemory`] = float64(v.Total)
	c.gauge[`FreeMemory`] = float64(v.Free)

	for i := 0; i < len(cp); i++ {
		c.gauge[`CPUutilization`+strconv.Itoa(i)] = float64(cp[i].System)
	}
}

type Statuses struct {
	Success  int
	Internal int
	Fail     int
}

func (c *Collector) sendMetrics() {
	c.Lock()
	defer c.Unlock()

	// Create a channel to send metrics
	metric := make(chan Metric)

	jobs := len(c.gauge) + len(c.counter)
	rate := jobs

	if *RateLimit > 0 {
		rate = *RateLimit
	}

	// Launch workers
	for i := 0; i < rate; i++ {
		zap.L().Debug(`Metric worker launched`, zap.Int(`id`, i))
		go c.sendMetricWorker(i, metric)
	}

	// Add gauge metrics
	for i, v := range c.gauge {
		metric <- Metric{
			ID:    i,
			MType: `gauge`,
			Value: &v,
		}
	}

	// Add counter metrics
	for i, v := range c.counter {
		metric <- Metric{
			ID:    i,
			MType: `counter`,
			Delta: &v,
		}
	}

	// Wait for all metrics to be sent
	close(metric)
}

// sendMetricWorker is a function that processes metrics from a channel and sends them to a remote server.
//
// Parameters:
//   - id: an integer representing the worker's ID.
//   - metric: a channel that receives Metric objects.
func (c *Collector) sendMetricWorker(id int, metric <-chan Metric) {
	for m := range metric {
		var code = 0
		if err := c.retryIfError(
			func() error {
				c, err := c.sendPOST(m, nil)
				code = c
				return err
			},
		); err != nil {
			zap.L().Error(err.Error())
		}

		zap.L().Debug(`Metric worker finished`, zap.Int(`id`, id), zap.Int(`code`, code), zap.String(`id`, m.ID))
	}
}

// sendPOST sends a POST request to the server with the given metric.
//
// Parameters:
// - metric: the metric to be sent
func (c *Collector) sendPOST(metrics Metric, statuses *Statuses) (int, error) {
	b, err := json.Marshal(metrics)
	if err != nil {
		return 0, err
	}

	var gz bytes.Buffer

	// Compress the request body
	w := gzip.NewWriter(&gz)
	w.Write(b)
	w.Close()

	// Create the request
	req, err := http.NewRequest(http.MethodPost, `http://`+*ServerAddress+`/update/`, &gz)
	if err != nil {
		return 0, err
	}

	// Set headers
	req.Header.Set(`Content-Encoding`, `gzip`)
	req.Header.Set(`Accept-Encoding`, `gzip`)
	req.Header.Set(`Content-Type`, `application/json`)

	// Add hash header
	if *Key != `` {
		c.addHashHeader(req, gz.Bytes())
	}

	// Send the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	return res.StatusCode, nil
}

// addHashHeader adds a hash header to the given http.Request and sets the value of the 'HashSHA256' header field.
//
// Parameters:
// - req: a pointer to an http.Request object to which the hash header will be added.
// - body: a byte slice representing the body of the request.
func (c *Collector) addHashHeader(req *http.Request, body []byte) {
	key := sha256.Sum256([]byte(*Key))

	// Create cipher block
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		zap.L().Error(`Cannot create AES block`, zap.Error(err))
		return
	}

	// Create GCM
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		zap.L().Error(`Cannot create AES GCM`, zap.Error(err))
		return
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	// Encode body
	h := aesgcm.Seal(nil, nonce, body, nil)
	req.Header.Set(`HashSHA256`, hex.EncodeToString(h[:]))
}

// retryIfError retries the given function if it returns an error.
func (c *Collector) retryIfError(f func() error) error {
	return retry.Do(
		func() error {
			return f()
		},
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			timer := 1 + (n * 2)
			return time.Duration(timer) * time.Second
		}),
		retry.Attempts(3),
	)
}
