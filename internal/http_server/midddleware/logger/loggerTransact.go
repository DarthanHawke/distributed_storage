package logger

import (
	trlog "distributed_storage/internal/logger/transaction_logger"
	"distributed_storage/internal/storage"
	"fmt"
)

func NewTransactionLog(store storage.StoreUsage) (*trlog.FileTransactionLogger, error) {
	var err error
	transactionlogger, err := trlog.NewFileTransactionLogger("transaction.log")
	if err != nil {
		return nil, fmt.Errorf("failed to create event logger: %w", err)
	}

	events, errors := transactionlogger.ReadEvents()
	count, ok, e := 0, true, trlog.TransactionEvent{}
	for ok && err == nil {
		select {
		case err, ok = <-errors: // Получает ошибки
		case e, ok = <-events:
			switch e.EventType {
			case trlog.EventDelete: // Получено событие DELETE!
				err = store.DeleteData(e.Key)
				count++
			case trlog.EventPut: // Получено событие PUT!
				err = store.PutData(e.Key, e.Value)
				count++
			}
		}
	}

	return transactionlogger, err
}
