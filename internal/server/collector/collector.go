package collector

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
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

func CollectMetric(s storage.Storage) *CollectorHandler {
	return &CollectorHandler{
		storage: s,
	}
}

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

func (c *CollectorHandler) parseURL(urlString string) (*ParsedURL, error) {
	// Prepare for .env
	endpoint := `/update/`

	// Remove prefix
	after, found := strings.CutPrefix(urlString, endpoint)
	if !found {
		fmt.Println(errNotFound.Error(), `on`, urlString)
		return nil, errNotFound
	}

	// Split url
	u := strings.Split(after, `/`)
	if len(u) != 3 {
		fmt.Println(errNotFound.Error(), `on`, urlString)
		return nil, errNotFound
	}

	// Replace %20 with space and check for empty
	for i := 0; i < len(u); i++ {
		if u[i] == "" {
			fmt.Println(errNotFound.Error(), `on`, urlString)
			return nil, errNotFound
		}
		u[i] = strings.Replace(u[i], `%20`, ``, -1)
	}

	return &ParsedURL{
		Type:  u[0],
		Name:  u[1],
		Value: u[2],
	}, nil
}

func (c *CollectorHandler) parseCounter(param string) (int64, error) {
	// If param is empty
	if param == `` {
		fmt.Println(errParse.Error(), `on`, param)
		return 0, errParse
	}

	// Parse
	n, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		fmt.Println(errParse.Error(), `on`, param)
		return 0, errParse
	}

	return n, nil
}

func (c *CollectorHandler) parseGauge(param string) (float64, error) {
	// If param is empty
	if param == `` {
		fmt.Println(errParse.Error(), `on`, param)
		return 0, errParse
	}

	// Parse
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		fmt.Println(errParse.Error(), `on`, param)
		return 0, errParse
	}

	return n, nil
}
