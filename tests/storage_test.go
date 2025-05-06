package tests

import (
	"distributed_storage/internal/storage"
	"errors"
	"testing"
)

func TestPut(t *testing.T) {
	const key = "create-key"
	const value = "create-value"

	var val interface{}
	var contains bool

	store, err := storage.New()
	if err != nil {
		t.Error("Faild to init storage")
	}

	defer delete(store.MyMap, key)

	// Sanity check
	_, contains = store.MyMap[key]
	if contains {
		t.Error("key/value already exists")
	}

	// err should be nil
	err = store.PutData(key, value)
	if err != nil {
		t.Error(err)
	}

	val, contains = store.MyMap[key]
	if !contains {
		t.Error("create failed")
	}

	if val != value {
		t.Error("val/value mismatch")
	}
}

func TestGet(t *testing.T) {
	const key = "read-key"
	const value = "read-value"

	var val interface{}
	var err error

	store, err := storage.New()
	if err != nil {
		t.Error("Faild to init storage")
	}

	defer delete(store.MyMap, key)

	// Read a non-thing
	val, err = store.GetData(key)
	if err == nil {
		t.Error("expected an error")
	}
	if !errors.Is(err, storage.ErrorNoSuchKey) {
		t.Error("unexpected error:", err)
	}

	store.MyMap[key] = value

	val, err = store.GetData(key)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	if val != value {
		t.Error("val/value mismatch")
	}
}

func TestDelete(t *testing.T) {
	const key = "delete-key"
	const value = "delete-value"

	var contains bool

	store, err := storage.New()
	if err != nil {
		t.Error("Faild to init storage")
	}

	defer delete(store.MyMap, key)

	store.MyMap[key] = value

	_, contains = store.MyMap[key]
	if !contains {
		t.Error("key/value doesn't exist")
	}

	store.DeleteData(key)

	_, contains = store.MyMap[key]
	if contains {
		t.Error("Delete failed")
	}
}
