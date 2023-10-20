package collector

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ParsedURL struct {
	Type  string
	Name  string
	Value string
}

type CollectorHandler struct {
	storage storage.Storage
}

var errNotFound error = errors.New(`404 page not found`)
var errType error = errors.New(`type not found`)
var errParse = errors.New(`parse error`)

// Create a new instance of CollectorHandler with the provided storage.
//
// It takes a storage.Storage object as a parameter and returns a pointer to a CollectorHandler.
func CollectMetric(s storage.Storage) *CollectorHandler {
	return &CollectorHandler{
		storage: s,
	}
}

// Handle the HTTP request and updates the corresponding metric.
//
// It parses the URL to extract the necessary information for updating the metric.
// The URL should include the metric type (`counter` or `gauge`), the metric name, and the metric value.
// If the URL is not valid or the metric type is not supported, it returns an error response.
// It then updates the metric based on the parsed information.
// For `counter` type, it parses the counter value and updates the counter metric.
// For `gauge` type, it parses the gauge value and updates the gauge metric.
// Finally, it sets the response status to OK and returns the response.
func (c *CollectorHandler) ServeHTTP(ctx *gin.Context) {
	// Parse URL
	parsedURL, err := c.parseURL(ctx.Request.URL.String())
	if err != nil {
		if err.Error() == errNotFound.Error() {
			ctx.String(http.StatusNotFound, err.Error())
		} else {
			ctx.String(http.StatusBadRequest, err.Error())
		}
		return
	}

	// Check type
	if parsedURL.Type != `counter` && parsedURL.Type != `gauge` {
		ctx.String(http.StatusBadRequest, errType.Error())
		return
	}

	// Update
	if parsedURL.Type == `counter` {
		value, err := c.parseCounter(parsedURL.Value)
		if err != nil {
			ctx.String(http.StatusBadRequest, err.Error())
			return
		}
		c.storage.UpdateCounterMetric(parsedURL.Name, value)

	} else if parsedURL.Type == `gauge` {
		value, err := c.parseGauge(parsedURL.Value)
		if err != nil {
			ctx.String(http.StatusBadRequest, err.Error())
			return
		}
		c.storage.UpdateGaugeMetric(parsedURL.Name, value)
	}

	// Response
	ctx.Header(`Content-Type`, `plain/text`)
	ctx.Status(http.StatusOK)
}

// Parse the given URL string and returns a parsed URL object and an error.
func (c *CollectorHandler) parseURL(urlString string) (*ParsedURL, error) {
	decodedURL, err := url.PathUnescape(urlString)
	if err != nil {
		zap.L().Error(err.Error())
	}

	endpoint := `/update/`

	// Remove prefix
	after, found := strings.CutPrefix(decodedURL, endpoint)
	if !found {
		zap.L().Error(errNotFound.Error() + ` on ` + urlString)
		return nil, errNotFound
	}

	// Split url
	u := strings.Split(after, `/`)
	if len(u) != 3 {
		zap.L().Error(errNotFound.Error() + ` on ` + urlString)
		return nil, errNotFound
	}

	// Check for empty and trim name
	for i := 0; i < len(u); i++ {
		if u[i] == "" {
			zap.L().Error(errNotFound.Error() + ` on ` + urlString)
			return nil, errNotFound
		}
		u[i] = strings.Trim(u[i], ` `)
	}

	return &ParsedURL{
		Type:  u[0],
		Name:  u[1],
		Value: u[2],
	}, nil
}

// Parse the given parameter and returns an int64 value.
//
// The param parameter is the string that needs to be parsed.
// The function returns an int64 value and an error.
func (c *CollectorHandler) parseCounter(param string) (int64, error) {
	// If param is empty
	if param == `` {
		zap.L().Error(errParse.Error() + ` on ` + param)
		return 0, errParse
	}

	// Parse
	n, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		zap.L().Error(errParse.Error() + ` on ` + param)
		return 0, errParse
	}

	return n, nil
}

// Parses a gauge value from a string.
//
// It takes a string parameter `param` and returns a float64 and an error.
func (c *CollectorHandler) parseGauge(param string) (float64, error) {
	// If param is empty
	if param == `` {
		zap.L().Error(errParse.Error() + ` on ` + param)
		return 0, errParse
	}

	// Parse
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		zap.L().Error(errParse.Error() + ` on ` + param)
		return 0, errParse
	}

	return n, nil
}
