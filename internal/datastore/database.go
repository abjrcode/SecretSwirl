package datastore

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type AppStore struct {
	dbConnectionString string
	dbFilePath         string
	inMemory           bool
}

func New(appDataDir, dbFileName string) *AppStore {
	runner := &AppStore{}

	dbFilePath := filepath.Join(appDataDir, dbFileName)
	dbFilePath = strings.ReplaceAll(dbFilePath, "\\", "/")

	runner.dbFilePath = dbFilePath
	runner.dbConnectionString = fmt.Sprintf("file:%s", dbFilePath)

	runner.inMemory = false

	return runner
}

func NewInMemory(dbFileName string) *AppStore {
	runner := &AppStore{}

	runner.dbConnectionString = fmt.Sprintf("file:%s?cache=shared&mode=memory", dbFileName)

	runner.inMemory = true

	return runner
}

func (store *AppStore) Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", store.dbConnectionString)

	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}

func copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func (store *AppStore) TakeBackup() error {
	if store.inMemory {
		return nil
	}

	return copyFile(store.dbFilePath, fmt.Sprintf("%s.bak", store.dbFilePath))
}

func (store *AppStore) RestoreBackup() error {
	if store.inMemory {
		return nil
	}

	backupFileName := fmt.Sprintf("%s.bak", store.dbFilePath)
	return copyFile(backupFileName, store.dbFilePath)
}
