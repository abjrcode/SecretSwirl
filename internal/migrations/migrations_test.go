package migrations

import (
	"io/fs"
	"testing"

	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type mockFS struct {
	MemFs afero.Fs
}

func (m *mockFS) Open(name string) (fs.File, error) {
	return m.MemFs.Open(name)
}

func TestNew(t *testing.T) {
	memFs := afero.NewMemMapFs()

	memFs.MkdirAll("migrations", 0755)
	afero.WriteFile(memFs, "/migrations/1_init.sql", []byte("CREATE TABLE dummy (id INTEGER PRIMARY KEY);"), 0644)

	mockFs := &mockFS{
		MemFs: memFs,
	}

	runner, err := New(mockFs, "migrations", nil, zerolog.Logger{}, testhelpers.NewMockErrorHandler(t))

	require.NoError(t, err)
	require.NotNil(t, runner)
}

func TestNewInvalidMigrationsPath(t *testing.T) {
	mockFs := &mockFS{
		MemFs: afero.NewMemMapFs(),
	}

	_, err := New(mockFs, "invalid-migrations-path", nil, zerolog.Logger{}, testhelpers.NewMockErrorHandler(t))

	require.Error(t, err)
}

func TestRunIfNecessary(t *testing.T) {
	dir := t.TempDir()
	memFs := afero.NewMemMapFs()

	memFs.MkdirAll("/migrations", 0755)
	afero.WriteFile(memFs, "/migrations/1_init.up.sql", []byte("CREATE TABLE dummy (id INTEGER PRIMARY KEY);"), 0644)
	afero.WriteFile(memFs, "/migrations/1_init.down.sql", []byte("DROP TABLE IF EXISTS dummy;"), 0644)
	afero.WriteFile(memFs, "/migrations/2_users.up.sql", []byte("CREATE TABLE users (id INTEGER PRIMARY KEY);"), 0644)
	afero.WriteFile(memFs, "/migrations/2_users.down.sql", []byte("DROP TABLE IF EXISTS users;"), 0644)

	mockFs := &mockFS{
		MemFs: memFs,
	}

	dataStore := datastore.New(dir, "test.db")
	runner, err := New(mockFs, "/migrations", dataStore, zerolog.Logger{}, testhelpers.NewMockErrorHandler(t))

	require.NoError(t, err)
	require.NotNil(t, runner)

	err = runner.RunSafe()

	require.NoError(t, err)

	db, err := dataStore.Open()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	_, err = db.Exec("SELECT * FROM dummy")
	require.NoError(t, err)

	_, err = db.Exec("SELECT * FROM users")
	require.NoError(t, err)
}

func TestRunIfNecessaryNoMigrationsNeeded(t *testing.T) {
	dir := t.TempDir()
	memFs := afero.NewMemMapFs()

	memFs.MkdirAll("/migrations", 0755)
	afero.WriteFile(memFs, "/migrations/1_init.up.sql", []byte("CREATE TABLE dummy (id INTEGER PRIMARY KEY);"), 0644)

	mockFs := &mockFS{
		MemFs: memFs,
	}

	dataStore := datastore.New(dir, "test.db")
	runner, err := New(mockFs, "/migrations", dataStore, zerolog.Logger{}, testhelpers.NewMockErrorHandler(t))

	require.NoError(t, err)
	require.NotNil(t, runner)

	err = runner.RunSafe()
	require.NoError(t, err)

	err = runner.RunSafe()

	require.NoError(t, err)
}

func TestRunIfNecessaryFailedMigration(t *testing.T) {
	memFs := afero.NewMemMapFs()

	memFs.MkdirAll("/migrations", 0755)
	afero.WriteFile(memFs, "/migrations/1_init.up.sql", []byte("bogus seeqel"), 0644)

	mockFs := &mockFS{
		MemFs: memFs,
	}

	dataStore := datastore.NewInMemory("test.db")
	runner, err := New(mockFs, "/migrations", dataStore, zerolog.Logger{}, testhelpers.NewMockErrorHandler(t))

	require.NoError(t, err)
	require.NotNil(t, runner)

	err = runner.RunSafe()

	require.Error(t, err)
}

func TestRunIfNecessaryFailedMigrationRestoresDatabase(t *testing.T) {
	dir := t.TempDir()
	memFs := afero.NewMemMapFs()

	memFs.MkdirAll("/migrations", 0755)
	afero.WriteFile(memFs, "/migrations/1_init.up.sql", []byte("CREATE TABLE dummy (id INTEGER PRIMARY KEY);"), 0644)

	mockFs := &mockFS{
		MemFs: memFs,
	}

	dataStore := datastore.New(dir, "test.db")
	db, err := dataStore.Open()
	require.NoError(t, err)
	_, err = db.Exec("SELECT * FROM dummy")
	require.Error(t, err)
	db.Close()

	runner, err := New(mockFs, "/migrations", dataStore, zerolog.Logger{}, testhelpers.NewMockErrorHandler(t))

	require.NoError(t, err)
	require.NotNil(t, runner)

	err = runner.RunSafe()
	require.NoError(t, err)

	db, err = dataStore.Open()
	require.NoError(t, err)
	_, err = db.Exec("SELECT * FROM dummy")
	require.NoError(t, err)
	db.Close()

	// For some reason the file system is getting wiped after first migration run
	// so we need to recreate the runner with a new instance seeded with the same migrations
	memFs = afero.NewMemMapFs()

	memFs.MkdirAll("/migrations", 0755)
	afero.WriteFile(memFs, "/migrations/1_init.up.sql", []byte("CREATE TABLE dummy (id INTEGER PRIMARY KEY);"), 0644)
	afero.WriteFile(memFs, "/migrations/2_semi_bogus.up.sql", []byte("CREATE TABLE users (id INTEGER PRIMARY KEY); bogus;"), 0644)

	mockFs = &mockFS{
		MemFs: memFs,
	}

	runner, err = New(mockFs, "/migrations", dataStore, zerolog.Logger{}, testhelpers.NewMockErrorHandler(t))
	require.NoError(t, err)

	err = runner.RunSafe()

	require.Error(t, err)

	db, err = dataStore.Open()
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})

	_, err = db.Exec("SELECT * FROM dummy")
	require.NoError(t, err)

	_, err = db.Exec("SELECT * FROM users")
	require.Error(t, err)
}
