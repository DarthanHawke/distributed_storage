package transactionlogger

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"sync"
)

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error
	ReadEvents() (<-chan TransactionEvent, <-chan error)
	Run()
	Wait()
	Close() error
}

type FileTransactionLogger struct {
	Events       chan<- TransactionEvent // Канал только для записи; для передачи событий
	Errors       <-chan error            // Канал только для чтения; для приема ошибок
	LastSequence uint64                  // Последний использованный порядковый номер
	file         *os.File                // Местоположение файла журнала
	wg           *sync.WaitGroup
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.wg.Add(1)
	l.Events <- TransactionEvent{EventType: EventPut, Key: key, Value: url.QueryEscape(value)}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.wg.Add(1)
	l.Events <- TransactionEvent{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.Errors
}

func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	var err error
	var l FileTransactionLogger = FileTransactionLogger{wg: &sync.WaitGroup{}}

	// Open the transaction log file for reading and writing.
	l.file, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &l, nil
}

func (l *FileTransactionLogger) Run() {
	events := make(chan TransactionEvent, 16) // Создать канал событий
	l.Events = events

	errors := make(chan error, 1) // Создать канал ошибок
	l.Errors = errors

	go func() {
		for e := range events { // Извлечь следующее событие Event
			l.LastSequence++ // Увеличить порядковый номер

			_, err := fmt.Fprintf(l.file, "%d\t%d\t%s\t%s\n", l.LastSequence, e.EventType, e.Key, e.Value) // Записать событие в журнал

			if err != nil {
				errors <- fmt.Errorf("cannot write to log file: %w", err)
				return
			}

			l.wg.Done()
		}
	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan TransactionEvent, <-chan error) {
	scanner := bufio.NewScanner(l.file)     // Создать Scanner для чтения l.file
	outEvent := make(chan TransactionEvent) // Небуферизованный канал событий
	outError := make(chan error, 1)         // Буферизованный канал ошибок

	go func() {
		var e TransactionEvent

		defer close(outEvent) // Закрыть каналы
		defer close(outError) // по завершении сопрограммы

		for scanner.Scan() {
			line := scanner.Text()

			fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value)
			// Проверка целостности!
			// Порядковые номера последовательно увеличиваются?
			if l.LastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			uv, err := url.QueryUnescape(e.Value)
			if err != nil {
				outError <- fmt.Errorf("value decoding failure: %w", err)
				return
			}

			e.Value = uv

			l.LastSequence = e.Sequence // Запомнить последний использованный
			// порядковый номер
			outEvent <- e // Отправить событие along
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}

func (l *FileTransactionLogger) Wait() {
	l.wg.Wait()
}

func (l *FileTransactionLogger) Close() error {
	l.Wait()

	if l.Events != nil {
		close(l.Events) // Terminates Run loop and goroutine
	}

	return l.file.Close()
}
