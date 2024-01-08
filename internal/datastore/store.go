package datastore

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/abjrcode/swervo/internal/utils"
)

type AppStore interface {
	GetDbFilePath() string

	Open() (*sql.DB, error)
	Close(*sql.DB) error

	TakeBackup() error
	RestoreBackup() error
}

type appStore struct {
	dbConnectionString string
	dbFilePath         string
}

func New(appDataDir, dbFileName string) AppStore {
	runner := &appStore{}

	dbFilePath := filepath.Join(appDataDir, dbFileName)
	dbFilePath = strings.ReplaceAll(dbFilePath, "\\", "/")

	runner.dbFilePath = dbFilePath
	runner.dbConnectionString = fmt.Sprintf("file:%s", dbFilePath)

	return runner
}

func (store *appStore) GetDbFilePath() string {
	return store.dbFilePath
}

func (store *appStore) Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", store.dbConnectionString)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(4)
	db.SetConnMaxIdleTime(60 * time.Second)
	db.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

func (store *appStore) Close(db *sql.DB) error {
	return db.Close()
}

func (store *appStore) TakeBackup() error {
	return utils.CopyFile(store.dbFilePath, fmt.Sprintf("%s.bak", store.dbFilePath))
}

func (store *appStore) RestoreBackup() error {
	backupFileName := fmt.Sprintf("%s.bak", store.dbFilePath)
	return utils.CopyFile(backupFileName, store.dbFilePath)
}
