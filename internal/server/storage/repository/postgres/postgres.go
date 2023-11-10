package postgres

import (
	"flag"
	"os"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var (
	PostgresDSN = flag.String(`d`, `postgres://user:password@postgres/metrics`, `Postgres DSN`)
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

// envParse initializes the StoreInterval, FileStoragePath, and Restore
// variables by checking for corresponding environment variables.
func envParse() {
	if env, exist := os.LookupEnv(`DATABASE_DSN `); exist {
		PostgresDSN = &env
	}
}

type PostgresStorage struct {
	db *sqlx.DB
}

// CreateRepository creates a new storage repository.
//
// Returns:
// - a pointer to a storage.Storage interface.
func CreateRepository() storage.Storage {
	// Parse environment variables
	envParse()

	db, err := sqlx.Connect(`postgres`, *PostgresDSN)
	if err != nil {
		zap.L().Error(err.Error())
	}

	db.MustExec(schema)

	return &PostgresStorage{
		db: db,
	}
}

// StartTickers a not used here
func (r *PostgresStorage) StartTickers() {}

// GetValues returns the gauge and counter maps of the MemStorage.
//
// Returns:
// - map[string]float64
// - map[string]int64.
func (r *PostgresStorage) GetValues() (map[string]float64, map[string]int64) {
	gaugeModels := []GaugeModel{}
	counterModels := []CounterModel{}

	// Request gauge models
	err := r.db.Select(&gaugeModels, `SELECT * FROM gauge`)
	if err != nil {
		zap.L().Error(err.Error())
	}

	// Request counter models
	err = r.db.Select(&counterModels, `SELECT * FROM counter`)
	if err != nil {
		zap.L().Error(err.Error())
	}

	// Convert gauge models to maps
	gauge := make(map[string]float64)
	for _, model := range gaugeModels {
		gauge[model.Name] = model.Value
	}

	// Convert counter models to maps
	counter := make(map[string]int64)
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
	err := r.db.Get(&counterModel, `SELECT * FROM counter WHERE name = $1`, name)
	if err != nil {
		zap.L().Error(err.Error())
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
	err := r.db.Get(&gaugeModel, `SELECT * FROM gauge WHERE name = $1`, name)
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}

	return &gaugeModel, nil
}

// GetCounterValue retrieves the value of a counter by its name from the MemStorage.
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
		zap.L().Error(err.Error())
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
		// Insert
		_, err := r.db.NamedExec(`INSERT INTO counter (name, value) VALUES (:name, :value)`, CounterModel{Name: name, Value: value})
		if err != nil {
			zap.L().Error(err.Error())
			return 0
		}
	} else {
		// Update
		_, err := r.db.NamedExec(`UPDATE counter SET value = :value WHERE name = :name`, CounterModel{Name: name, Value: counterModel.Value + value})
		if err != nil {
			zap.L().Error(err.Error())
			return 0
		}
	}

	// Get updated value
	updatedCounterModel, err := r.GetCounterByName(name)
	if err != nil {
		zap.L().Error(err.Error())
		return 0
	}

	return updatedCounterModel.Value
}

// GetGaugeValue retrieves the value of a gauge by its name from the MemStorage.
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
		zap.L().Error(err.Error())
		return 0, false
	}

	return gaugeModel.Value, true
}

// UpdateGaugeMetric updates the gauge metric with the given name and value in the MemStorage.
//
// Parameters:
// - name: the name of the gauge metric (string)
// - value: the value of the gauge metric (float64)
//
// Returns:
// - the updated value of the gauge metric (float64).
func (r *PostgresStorage) UpdateGaugeMetric(name string, value float64) float64 {
	_, err := r.GetGaugeByName(name)

	// If counter doesn't exist
	if err != nil {
		// Insert
		_, err := r.db.NamedExec(`INSERT INTO gauge (name, value) VALUES (:name, :value)`, GaugeModel{Name: name, Value: value})
		if err != nil {
			zap.L().Error(err.Error())
			return 0
		}
	} else {
		// Update
		_, err := r.db.NamedExec(`UPDATE gauge SET value = :value WHERE name = :name`, GaugeModel{Name: name, Value: value})
		if err != nil {
			zap.L().Error(err.Error())
			return 0
		}
	}

	// Get updated value
	updatedGaugeModel, err := r.GetGaugeByName(name)
	if err != nil {
		zap.L().Error(err.Error())
		return 0
	}

	return updatedGaugeModel.Value
}
