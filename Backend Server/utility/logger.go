package utility

import (
	"io"
	"strconv"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(output io.Writer) *Logger {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true

	logger := logrus.New()
	logger.SetFormatter(customFormatter)
	logger.SetOutput(output)

	return &Logger{Logger: logger}
}

type FlatDatabaseWriter struct {
	db *badger.DB
}

func NewFlatDatabaseWriter(path string) (*FlatDatabaseWriter, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &FlatDatabaseWriter{db: db}, nil
}

func (w *FlatDatabaseWriter) Write(data []byte) (n int, err error) {
	err = w.db.Update(func(txn *badger.Txn) error {
		currentTimeString := strconv.FormatInt(time.Now().UnixNano(), 10)
		return txn.Set([]byte(currentTimeString), data)
	})
	
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (w *FlatDatabaseWriter) Close() error {
	return w.db.Close()
}
