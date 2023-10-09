package collector

import (
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
	gauge   map[string]float64
	counter map[string]int64
	done    chan struct{}
}

func init() {
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
}

// CreateCollector creates a new instance of the Collector struct.
//
// Returns a pointer to a Collector.
func CreateCollector() *Collector {
	return &Collector{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
		done:    make(chan struct{}),
	}
}

// StartTickers starts the tickers for collecting and sending metrics in the Collector struct.
func (c *Collector) StartTickers() {
	zap.L().Debug(fmt.Sprintf(`Poll Interval: %d`, PollInterval))
	zap.L().Debug(fmt.Sprintf(`Report Interval: %d`, ReportInterval))

	// Start tickers
	collectTicker := time.NewTicker(time.Duration(*PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(*ReportInterval) * time.Second)

	zap.L().Info(`Collector's tickers started`)

	for {
		select {
		case <-collectTicker.C:
			c.collectMetric()
		case <-sendTicker.C:
			c.sendMetrics()
		}
	}
}

// StopTickers stops the tickers of the Collector.
//
// It closes the 'done' channel and prints a message to the console.
func (c *Collector) StopTickers() {
	zap.L().Info(`Collector's tickers stopped`)
	close(c.done)
}

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

func (c *Collector) sendMetrics() {
	for name, value := range c.gauge {
		c.sendPOST(`gauge`, name, fmt.Sprintf(`%f`, value))
	}

	for name, value := range c.counter {
		c.sendPOST(`counter`, name, fmt.Sprintf(`%d`, value))
	}

	zap.L().Debug(`Metrics sended`)
}

func (c *Collector) sendPOST(metricType string, name string, value string) {
	res, err := http.Post(`http://`+*ServerAddress+`/update/`+metricType+`/`+name+`/`+value, `text/plain`, nil)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}
	res.Body.Close()
}
