package sqldb

import (
	"database/sql"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/lansfy/gonkex/storage/addons/sqldb/testfixtures"

	"github.com/stretchr/testify/require"
)

func Test_Storage_GetType(t *testing.T) {
	s := NewStorage(PostgreSQL, nil, nil)
	require.Equal(t, "postgresql", s.GetType())
}

func Test_Storage_InvalidDbType(t *testing.T) {
	s := NewStorage("", nil, nil)
	err := s.LoadFixtures("testdata", []string{"fixture.yaml"})
	require.Error(t, err)
	require.Equal(t, err.Error(), "unknown db type \"\"")

	_, err = s.ExecuteQuery("fake request")
	require.Error(t, err)
	require.Equal(t, err.Error(), "unknown db type \"\"")
}

func Test_Storage_LoadFixtures_InvalidPath(t *testing.T) {
	s := NewStorage(PostgreSQL, nil, nil)
	err := s.LoadFixtures("invalid/path", []string{"nonexistent.yml"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "load fixtures")
}

func normalize(data []byte) string {
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

func Test_Storage_LoadFixtures(t *testing.T) {
	var called bool
	oldFunc := createFixtureParams
	defer func() { createFixtureParams = oldFunc }()

	expectedContent, err := os.ReadFile("testdata/fixture_processed.yaml")
	require.NoError(t, err)

	createFixtureParams = func(dbType SQLType, db *sql.DB, fs fstest.MapFS) []func(*testfixtures.Loader) error {
		called = true
		require.Equal(t, PostgreSQL, dbType)
		require.Equal(t, normalize(expectedContent), normalize(fs[virtualFileName].Data))
		return nil
	}

	s := &Storage{dbType: PostgreSQL}
	err = s.LoadFixtures("testdata", []string{"fixture.yaml"})
	require.Error(t, err)
	require.Equal(t, err.Error(), "testfixtures: database is required")
	require.True(t, called)
}

func Test_createFixtureParams(t *testing.T) {
	opts := createFixtureParams(PostgreSQL, nil, fstest.MapFS{})
	require.Equal(t, 7, len(opts))
}
