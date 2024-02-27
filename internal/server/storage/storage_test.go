package storage_test

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
)

func Example() {
	// Create storage
	var s storage.Storage
	if storage, ok := repository.CreateRepository(); ok {
		s = storage
	} else {
		s = nil
	}

	// Use repository
	s.GetValues()
}
