// Package repository init env and init main storage repository
package repository

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository/memory"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository/postgres"
	"go.uber.org/zap"
)

var (
	StoreInterval     time.Duration
	PostgresDSN       = flag.String(`d`, ``, `Postgres DSN`)
	FileStoragePath   = flag.String(`f`, `/tmp/metrics-db.json`, `File storage path`)
	Restore           = flag.Bool(`r`, true, `Restore from file`)
	StoreIntervalFlag = flag.Int(`i`, 300, `Store interval in seconds`) // Cannot use flag.Duration because Yandex's autotest send int
)

type ServerConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

// envParse initializes the StoreInterval, FileStoragePath, and Restore
// variables by checking for corresponding environment variables.
func envParse() {
	if env, exist := os.LookupEnv(`DATABASE_DSN`); exist {
		PostgresDSN = &env
	}

	if env, exist := os.LookupEnv(`STORE_INTERVAL`); exist {
		if dur, err := time.ParseDuration(env); err == nil {
			StoreInterval = dur
		}
	} else {
		StoreInterval = time.Duration(*StoreIntervalFlag) * time.Second
	}

	if env, exist := os.LookupEnv(`FILE_STORAGE_PATH`); exist {
		FileStoragePath = &env
	}

	if env, exist := os.LookupEnv(`RESTORE`); exist {
		if b, err := strconv.ParseBool(env); err == nil {
			Restore = &b
		}
	}

	if file, err := os.Open(`./server.config.json`); err == nil {
		defer file.Close()

		b, _ := io.ReadAll(file)
		var config ServerConfig
		json.Unmarshal(b, &config)

		PostgresDSN = &config.DatabaseDSN
		Restore = &config.Restore
		FileStoragePath = &config.StoreFile
	}

}

// CreateRepository creates a storage object based on the provided configuration.
//
// This function parses the environment variables or flags, logs the created storage,
// and then creates and returns the appropriate storage object based on the configuration.
//
// Return:
// - The created storage object of type `storage.Storage`.
func CreateRepository() (storage.Storage, bool) {
	// Parse environment variables
	envParse()

	// Log created storage
	zap.L().Debug(`Storage parameters:`,
		zap.String(`PostgresDSN`, *PostgresDSN),
		zap.String(`FileStoragePath`, *FileStoragePath),
		zap.Duration(`StoreInterval`, StoreInterval),
		zap.Bool(`Restore`, *Restore),
	)

	// Create Postgres storage
	if *PostgresDSN != `` {
		zap.L().Debug(`PostgresStorage created`)
		p := postgres.CreateRepository(postgres.Options{
			PostgresDSN: PostgresDSN,
		})
		return p, p != nil
	}

	// Create memory storage
	zap.L().Debug(`MemStorage created`)
	memStorage := memory.CreateRepository(memory.Options{
		StoreInterval:   StoreInterval,
		FileStoragePath: FileStoragePath,
		Restore:         Restore,
	})

	// Start tickers for MemStorage
	memStorage.StartTickers()

	return memStorage, true
}
