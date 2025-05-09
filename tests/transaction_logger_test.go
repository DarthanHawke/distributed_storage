package tests

import (
	trlog "distributed_storage/internal/logger/transaction_logger"
	"os"
	"testing"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func TestCreateLogger(t *testing.T) {
	const filename = "tmp/create-logger.txt"
	defer os.Remove(filename)

	tl, err := trlog.NewFileTransactionLogger(filename)

	if tl == nil {
		t.Error("Logger is nil?")
	}

	if err != nil {
		t.Errorf("Got error: %v", err)
	}

	if !fileExists(filename) {
		t.Errorf("File %s doesn't exist", filename)
	}
}

func TestWriteAppend(t *testing.T) {
	const filename = "tmp/create-logger.txt"
	defer os.Remove(filename)

	tl, err := trlog.NewFileTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}
	tl.Run()
	defer tl.Close()

	chev, cherr := tl.ReadEvents()
	for e := range chev {
		t.Log(e)
	}
	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	tl.WritePut("my-key", "my-value")
	tl.WritePut("my-key", "my-value2")
	tl.Wait()

	tl2, err := trlog.NewFileTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}
	tl2.Run()
	defer tl2.Close()

	chev, cherr = tl2.ReadEvents()
	for e := range chev {
		t.Log(e)
	}
	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	tl2.WritePut("my-key", "my-value3")
	tl2.WritePut("my-key2", "my-value4")
	tl2.Wait()

	if tl2.LastSequence != 4 {
		t.Errorf("Last sequence mismatch (expected 4; got %d)", tl2.LastSequence)
	}
}

func TestWritePut(t *testing.T) {
	const filename = "tmp/create-logger.txt"
	defer os.Remove(filename)

	tl, _ := trlog.NewFileTransactionLogger(filename)
	tl.Run()
	defer tl.Close()

	tl.WritePut("my-key", "my-value")
	tl.WritePut("my-key", "my-value2")
	tl.WritePut("my-key", "my-value3")
	tl.WritePut("my-key", "my-value4")
	tl.Wait()

	tl2, _ := trlog.NewFileTransactionLogger(filename)
	evin, errin := tl2.ReadEvents()
	defer tl2.Close()

	for e := range evin {
		t.Log(e)
	}

	err := <-errin
	if err != nil {
		t.Error(err)
	}

	if tl.LastSequence != tl2.LastSequence {
		t.Errorf("Last sequence mismatch (%d vs %d)", tl.LastSequence, tl2.LastSequence)
	}
}
