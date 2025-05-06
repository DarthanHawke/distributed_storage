package storage

import (
	"errors"
	"sync"
)

type StoreUsage interface {
	PutData(key string, value string) error
	GetData(key string) (string, error)
	DeleteData(key string) error
}

type Store struct {
	sync.RWMutex
	MyMap map[string]string
}

func New() (*Store, error) {
	return &Store{MyMap: make(map[string]string)}, nil
}

var ErrorNoSuchKey = errors.New("no such key")

func (store *Store) PutData(key string, value string) error {
	store.Lock()
	store.MyMap[key] = value
	store.Unlock()

	return nil
}

func (store *Store) GetData(key string) (string, error) {
	store.RLock()
	value, ok := store.MyMap[key]
	store.RUnlock()
	if !ok {
		return "", ErrorNoSuchKey
	}
	return value, nil
}

func (store *Store) DeleteData(key string) error {
	store.Lock()
	delete(store.MyMap, key)
	store.Unlock()

	return nil
}
