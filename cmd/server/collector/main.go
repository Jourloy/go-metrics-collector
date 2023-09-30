package collector

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jourloy/go-metrics-collector/cmd/server/storage"
)

type ParsedURL struct {
	Type  string
	Name  string
	Value string
}

type CollectorHandler struct {
	storage storage.Storage
}

var errPrefixError error = errors.New(`not found prefix`)
var errLengthError error = errors.New(`length of url params is not 3`)
var errEmptyError error = errors.New(`empty url params`)
var errMethodError error = errors.New(`method not allowed`)
var errTypeError error = errors.New(`type not found`)
var errParseError error = errors.New(`parse error`)

func CollectMetric(s storage.Storage) *CollectorHandler {
	return &CollectorHandler{
		storage: s,
	}
}

func (c *CollectorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Check method
	if r.Method != http.MethodPost {
		http.Error(w, errMethodError.Error(), http.StatusMethodNotAllowed)
		return
	}

	// Parse URL
	parsedURL, err := c.parseURL(r.URL.String())
	if err != nil {
		if err.Error() == `not found prefix` {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == `length of url params is not 3` {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == `empty url params` {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// Check type
	if parsedURL.Type != `counter` && parsedURL.Type != `gauge` {
		http.Error(w, errTypeError.Error(), http.StatusBadRequest)
		return
	}

	// Update
	if parsedURL.Type == `counter` {
		value, err := c.parseCounter(parsedURL.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.storage.UpdateCounterMetric(parsedURL.Name, value)

	} else if parsedURL.Type == `gauge` {
		value, err := c.parseGauge(parsedURL.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.storage.UpdateGaugeMetric(parsedURL.Name, value)
	}

	// Response
	w.Header().Set(`Content-Type`, `plain/text`)
	w.WriteHeader(http.StatusOK)
}

func (c *CollectorHandler) parseURL(urlString string) (*ParsedURL, error) {
	// Prepare for .env
	endpoint := `/update/`

	// Remove prefix
	after, found := strings.CutPrefix(urlString, endpoint)
	if !found {
		fmt.Println(errPrefixError.Error(), `on`, urlString)
		return nil, errPrefixError
	}

	// Split url
	u := strings.Split(after, `/`)
	if len(u) != 3 {
		fmt.Println(errLengthError.Error(), `on`, urlString)
		return nil, errLengthError
	}

	// Replace %20 with space and check for empty
	for i := 0; i < len(u); i++ {
		if u[i] == "" {
			fmt.Println(errEmptyError.Error(), `on`, urlString)
			return nil, errEmptyError
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
		fmt.Println(errParseError.Error(), `on`, param)
		return 0, errParseError
	}

	// Parse
	n, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		fmt.Println(errParseError.Error(), `on`, param)
		return 0, errParseError
	}

	return n, nil
}

func (c *CollectorHandler) parseGauge(param string) (float64, error) {
	// If param is empty
	if param == `` {
		fmt.Println(errParseError.Error(), `on`, param)
		return 0, errParseError
	}

	// Parse
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		fmt.Println(errParseError.Error(), `on`, param)
		return 0, errParseError
	}
	
	return n, nil
}
