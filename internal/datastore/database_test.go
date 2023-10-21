package datastore

import (
	"crypto/sha256"
	"database/sql"
	"io"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	dir := t.TempDir()

	dataStore := New(dir, "test.db")

	require.NotNil(t, dataStore)
}

func TestNewInMemory(t *testing.T) {
	dataStore := NewInMemory("test.db")

	require.NotNil(t, dataStore)
}

func TestOpen(t *testing.T) {
	dir := t.TempDir()

	dataStore := New(dir, "test.db")

	db, err := dataStore.Open()

	t.Cleanup(func() {
		db.Close()
	})

	require.NoError(t, err)
	require.NotNil(t, db)
	require.FileExists(t, dataStore.dbFilePath)
}

func TestOpenInMemory(t *testing.T) {
	dataStore := NewInMemory("test.db")

	db, err := dataStore.Open()

	t.Cleanup(func() {
		db.Close()
	})

	require.NoError(t, err)
	require.NotNil(t, db)
	require.Empty(t, dataStore.dbFilePath)
}

func calculateFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return string(hasher.Sum(nil)), nil
}

func seedDatabase(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO users (name) VALUES ('John Doe')`)
	require.NoError(t, err)
}

func TestTakeBackup(t *testing.T) {
	dir := t.TempDir()

	originalStore := New(dir, "test.db")

	db, err := originalStore.Open()
	require.NoError(t, err)
	seedDatabase(t, db)
	db.Close()

	originalHash, err := calculateFileHash(originalStore.dbFilePath)
	require.NoError(t, err)

	require.NotNil(t, db)

	err = originalStore.TakeBackup()

	require.NoError(t, err)
	require.FileExists(t, originalStore.dbFilePath+".bak")

	backupHash, err := calculateFileHash(originalStore.dbFilePath + ".bak")
	require.NoError(t, err)

	require.Equal(t, originalHash, backupHash)

	backupStore := New(dir, "test.db.bak")

	db, err = backupStore.Open()
	require.NoError(t, err)

	var name string
	err = db.QueryRow(`SELECT name FROM users WHERE id = 1`).Scan(&name)
	require.NoError(t, err)
}

func TestTakeBackupInMemory(t *testing.T) {
	dataStore := NewInMemory("test.db")

	err := dataStore.TakeBackup()

	require.NoError(t, err)
	require.NoFileExists(t, dataStore.dbFilePath)
	require.NoFileExists(t, dataStore.dbFilePath+".bak")
}

func TestRestoreBackup(t *testing.T) {
	dir := t.TempDir()

	originalStore := New(dir, "test.db")

	db, err := originalStore.Open()
	require.NoError(t, err)

	originalHash, err := calculateFileHash(originalStore.dbFilePath)
	require.NoError(t, err)

	require.NotNil(t, db)

	err = originalStore.TakeBackup()
	require.NoError(t, err)

	seedDatabase(t, db)
	db.Close()

	err = originalStore.RestoreBackup()
	require.NoError(t, err)

	require.FileExists(t, originalStore.dbFilePath)
	require.FileExists(t, originalStore.dbFilePath+".bak")

	backupHash, err := calculateFileHash(originalStore.dbFilePath)
	require.NoError(t, err)

	require.Equal(t, originalHash, backupHash)

	restoredStore := New(dir, "test.db")

	db, err = restoredStore.Open()
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})

	var name string
	err = db.QueryRow(`SELECT name FROM users WHERE id = 1`).Scan(&name)
	require.Error(t, err)
}

func TestRestoreBackupWithOpenConnections(t *testing.T) {
	dir := t.TempDir()

	dataStore := New(dir, "test.db")

	db, err := dataStore.Open()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	originalHash, err := calculateFileHash(dataStore.dbFilePath)
	require.NoError(t, err)

	require.NotNil(t, db)

	err = dataStore.TakeBackup()
	require.NoError(t, err)

	err = dataStore.RestoreBackup()
	require.NoError(t, err)

	require.FileExists(t, dataStore.dbFilePath)
	require.FileExists(t, dataStore.dbFilePath+".bak")

	backupHash, err := calculateFileHash(dataStore.dbFilePath)
	require.NoError(t, err)

	require.Equal(t, originalHash, backupHash)
}
