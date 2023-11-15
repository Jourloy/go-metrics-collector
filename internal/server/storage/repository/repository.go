package repository

import (
	"flag"
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
}

func CreateRepository() storage.Storage {
	// Parse environment variables
	envParse()

	// Log created storage
	zap.L().Debug(`Storage parameters:`,
		zap.String(`PostgresDSN`, *PostgresDSN),
		zap.String(`FileStoragePath`, *FileStoragePath),
		zap.Duration(`StoreInterval`, StoreInterval),
		zap.Bool(`Restore`, *Restore),
	)

	if *PostgresDSN != `` {
		zap.L().Debug(`PostgresStorage created`)
		return postgres.CreateRepository(postgres.Options{
			PostgresDSN: PostgresDSN,
		})
	}

	zap.L().Debug(`MemStorage created`)
	return memory.CreateRepository(memory.Options{
		StoreInterval:   StoreInterval,
		FileStoragePath: FileStoragePath,
		Restore:         Restore,
	})
}
