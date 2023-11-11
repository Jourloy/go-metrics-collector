package repository

import (
	"flag"
	"os"
	"strconv"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository/memory"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository/postgres"
	"go.uber.org/zap"
)

var (
	PostgresDSN     = flag.String(`d`, ``, `Postgres DSN`)
	StoreInterval   = flag.Int(`i`, 300, `Store interval in seconds`)
	FileStoragePath = flag.String(`f`, `/tmp/metrics-db.json`, `File storage path`)
	Restore         = flag.Bool(`r`, true, `Restore from file`)
)

// envParse initializes the StoreInterval, FileStoragePath, and Restore
// variables by checking for corresponding environment variables.
func envParse() {
	if env, exist := os.LookupEnv(`DATABASE_DSN`); exist {
		PostgresDSN = &env
	}

	if env, exist := os.LookupEnv(`STORE_INTERVAL`); exist {
		if i, err := strconv.Atoi(env); err == nil {
			StoreInterval = &i
		}
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
		zap.Int(`StoreInterval`, *StoreInterval),
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
