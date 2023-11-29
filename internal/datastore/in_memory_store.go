package datastore

import (
	"database/sql"
	"fmt"
)

type inMemoryAppStore struct {
	dbConnectionString string
	dbFilePath         string
	inMemoryDb         *sql.DB
}

func NewInMemory(dbFileName string) AppStore {
	store := &inMemoryAppStore{}

	store.dbConnectionString = fmt.Sprintf("file:%s?mode=memory&cache=shared", dbFileName)

	return store
}

func (store *inMemoryAppStore) GetDbFilePath() string {
	return store.dbFilePath
}

func (store *inMemoryAppStore) Open() (*sql.DB, error) {
	if store.inMemoryDb != nil {
		return store.inMemoryDb, nil
	}

	db, err := sql.Open("sqlite3", store.dbConnectionString)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (store *inMemoryAppStore) Close(db *sql.DB) error {
	/**
	 * This is a no-op.
	 * We don't close the in-memory database because:
	 * 1. They are only used for automated testing
	 * 2. Closing sqlite in-memory connection will delete the database
	 */
	return nil
}

func (store *inMemoryAppStore) TakeBackup() error {
	return nil
}

func (store *inMemoryAppStore) RestoreBackup() error {
	return nil
}
