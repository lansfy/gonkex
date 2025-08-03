package sqldb

import (
	"database/sql"
	"testing/fstest"

	"github.com/lansfy/gonkex/storage/addons/sqldb/testfixtures"
)

func LoadFixtures(dialect SQLType, db *sql.DB, location string, names []string) error {
	data, err := ConvertToTestFixtures(CreateFileLoader(location), names)
	if err != nil {
		return err
	}

	vfs := fstest.MapFS{
		virtualFileName: &fstest.MapFile{
			Data: data,
		},
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect(string(dialect)),
		testfixtures.FS(vfs),
		testfixtures.FilesMultiTables(virtualFileName),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.SkipTableChecksumComputation(),
		testfixtures.ResetSequencesTo(1),
	)
	if err != nil {
		return err
	}
	return fixtures.Load()
}
