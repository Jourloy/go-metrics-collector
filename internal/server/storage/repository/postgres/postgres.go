// Package postgres provide interface for store data in Postgres
package postgres

import (
	"slices"
	"time"

	"github.com/avast/retry-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var schema = `
CREATE TABLE IF NOT EXISTS gauge (
	name VARCHAR(255) PRIMARY KEY,
	value FLOAT
);

CREATE TABLE IF NOT EXISTS counter (
	name VARCHAR(255) PRIMARY KEY,
	value BIGINT
)`

type GaugeModel struct {
	Name  string  `db:"name"`
	Value float64 `db:"value"`
}

type CounterModel struct {
	Name  string `db:"name"`
	Value int64  `db:"value"`
}

type PostgresStorage struct {
	db *sqlx.DB
}

type Options struct {
	PostgresDSN *string
}

// CreateRepository creates a new storage repository.
//
// Returns:
// - a pointer to a storage.Storage interface.
func CreateRepository(opt Options) *PostgresStorage {
	var db *sqlx.DB

	// Connect to Postgres
	err := retryIfError(
		func() error {
			database, err := sqlx.Connect(`postgres`, *opt.PostgresDSN)
			if err != nil {
				zap.L().Error(err.Error())
				return err
			}
			db = database
			return nil
		},
	)
	if err != nil {
		zap.L().Error(err.Error())
		return nil
	}

	// Create tables
	db.MustExec(schema)

	return &PostgresStorage{
		db: db,
	}
}

// StartTickers a not used here
func (r *PostgresStorage) StartTickers() {}

// GetValues returns the gauge and counter maps of the postgres database.
//
// Returns:
// - map[string]float64
// - map[string]int64.
func (r *PostgresStorage) GetValues() (map[string]float64, map[string]int64) {
	gaugeModels := []GaugeModel{}
	counterModels := []CounterModel{}

	// Request gauge models
	if err := retryIfError(
		func() error {
			return r.db.Select(&gaugeModels, `SELECT name, value FROM gauge`)
		},
	); err != nil {
		zap.L().Error(err.Error())
		return nil, nil
	}

	// Request counter models
	if err := retryIfError(func() error {
		return r.db.Select(&counterModels, `SELECT name, value FROM counter`)
	}); err != nil {
		zap.L().Error(`Error while getting data from Postgres`, zap.Error(err))
	}

	// Convert gauge models to maps
	gauge := make(map[string]float64, len(gaugeModels))
	for _, model := range gaugeModels {
		gauge[model.Name] = model.Value
	}

	// Convert counter models to maps
	counter := make(map[string]int64, len(counterModels))
	for _, model := range counterModels {
		counter[model.Name] = model.Value
	}

	return gauge, counter
}

// GetCounterByName retrieves a CounterModel from the Postgres based on the given name.
//
// Parameters:
// - name: the name of the counter.
//
// Returns:
// - *CounterModel: a pointer to the CounterModel retrieved from the database.
// - error: any error that occurred during the retrieval process.
func (r *PostgresStorage) GetCounterByName(name string) (*CounterModel, error) {
	counterModel := CounterModel{}

	// Request counter model
	if err := retryIfError(func() error {
		return r.db.Get(&counterModel, `SELECT name, value FROM counter WHERE name = $1`, name)
	}); err != nil {
		zap.L().Error(`Error while operate with Postgres`, zap.Error(err))
		return nil, err
	}

	return &counterModel, nil
}

// GetGaugeByName retrieves a GaugeModel from the Postgres based on the given name.
//
// Parameters:
// - name: the name of the gauge.
//
// Returns:
// - *GaugeModel: a pointer to the GaugeModel retrieved from the database.
// - error: any error that occurred during the retrieval process.
func (r *PostgresStorage) GetGaugeByName(name string) (*GaugeModel, error) {
	gaugeModel := GaugeModel{}

	// Request counter model
	if err := retryIfError(func() error {
		return r.db.Get(&gaugeModel, `SELECT name, value FROM gauge WHERE name = $1`, name)
	}); err != nil {
		zap.L().Error(`Error while getting data from Postgres`, zap.Error(err))
		return nil, err
	}

	return &gaugeModel, nil
}

// GetCounterValue retrieves the value of a counter by its name from the postgres database.
//
// Parameters:
// - name: the name of the counter.
//
// Returns:
// - int64: the value of the counter.
// - bool: true if the counter exists, false otherwise.
func (r *PostgresStorage) GetCounterValue(name string) (int64, bool) {
	counterModel, err := r.GetCounterByName(name)
	if err != nil {
		return 0, false
	}

	return counterModel.Value, true
}

// UpdateCounterMetric updates the counter metric with the given name by adding the value to it.
//
// Parameters:
// - name: the name of the counter metric (string)
// - value: the value to be added to the counter metric (int64)
//
// Returns:
// - the updated value of the counter metric (int64)
func (r *PostgresStorage) UpdateCounterMetric(name string, value int64) int64 {
	counterModel, err := r.GetCounterByName(name)

	// If counter doesn't exist
	if err != nil {
		// Insert metric and return 0 if error
		if err := retryIfError(func() error {
			_, err := r.db.NamedExec(
				`INSERT INTO counter (name, value) VALUES (:name, :value)`,
				CounterModel{Name: name, Value: value},
			)
			return err
		}); err != nil {
			zap.L().Error(`Error while inserting data into Postgres`, zap.Error(err))
			return 0
		}
	} else {
		// Update metric and return 0 if error
		if err := retryIfError(func() error {
			_, err := r.db.NamedExec(
				`UPDATE counter SET value = :value WHERE name = :name`,
				CounterModel{Name: name, Value: counterModel.Value + value},
			)
			return err
		}); err != nil {
			zap.L().Error(`Error while updating data into Postgres`, zap.Error(err))
			return 0
		}
	}

	// Get updated value
	updatedCounterModel, err := r.GetCounterByName(name)
	if err != nil {
		return 0
	}

	return updatedCounterModel.Value
}

// GetGaugeValue retrieves the value of a gauge by its name from the postgres database.
//
// Parameters:
// - name: a string representing the name of the gauge.
//
// Returns:
// - value: a float64 representing the value of the gauge.
// - ok: a boolean indicating whether the gauge was found.
func (r *PostgresStorage) GetGaugeValue(name string) (float64, bool) {
	gaugeModel, err := r.GetGaugeByName(name)
	if err != nil {
		return 0, false
	}

	return gaugeModel.Value, true
}

// UpdateGaugeMetric updates the gauge metric with the given name and value in the postgres database.
//
// Parameters:
// - name: the name of the gauge metric (string)
// - value: the value of the gauge metric (float64)
//
// Returns:
// - the updated value of the gauge metric (float64).
func (r *PostgresStorage) UpdateGaugeMetric(name string, value float64) float64 {
	_, err := r.GetGaugeByName(name)

	// If gauge doesn't exist
	if err != nil {
		// Insert metric and return 0 if error
		if err := retryIfError(func() error {
			_, err := r.db.NamedExec(
				`INSERT INTO gauge (name, value) VALUES (:name, :value)`,
				GaugeModel{Name: name, Value: value},
			)
			return err
		}); err != nil {
			zap.L().Error(`Error while inserting data into Postgres`, zap.Error(err))
			return 0
		}
	} else {
		// Update metric and return 0 if error
		if err := retryIfError(func() error {
			_, err := r.db.NamedExec(
				`UPDATE gauge SET value = :value WHERE name = :name`,
				GaugeModel{Name: name, Value: value},
			)
			return err
		}); err != nil {
			zap.L().Error(`Error while updating data into Postgres`, zap.Error(err))
			return 0
		}
	}

	// Get updated value
	updatedGaugeModel, err := r.GetGaugeByName(name)
	if err != nil {
		return 0
	}

	return updatedGaugeModel.Value
}

// Class 08 errors
var retriableErrors = []string{
	`connection_exception`,
	`connection_does_not_exist`,
	`connection_failure`,
	`sqlclient_unable_to_establish_sqlconnection`,
	`sqlserver_rejected_establishment_of_sqlconnection`,
	`transaction_resolution_unknown`,
	`protocol_violation`,
}

// retryIfError retries the given function if it returns an error.
func retryIfError(f func() error) error {
	return retry.Do(
		func() error {
			return f()
		},
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			timer := 1 + (n * 2)
			return time.Duration(timer) * time.Second
		}),
		retry.Attempts(3),
		retry.RetryIf(func(err error) bool {
			return slices.Contains(retriableErrors, err.Error())
		}),
	)
}
