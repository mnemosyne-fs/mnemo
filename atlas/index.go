package atlas

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

var ()

var (
	//go:embed index_sql/init.sql
	init_sql string

	//go:embed index_sql/put.sql
	put_sql string

	//go:embed index_sql/select_path.sql
	select_path string

	//go:embed index_sql/delete_path.sql
	delete_path string
)

type IndexRow struct {
	path string
	size int
	hash string
}

type Index struct {
	db *sql.DB
}

func OpenIndex(file string) (*Index, error) {
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}

	return &Index{
		db: db,
	}, nil
}

func (i *Index) Init() error {
	_, err := i.db.Exec(init_sql)
	return err
}

// Will replace if path already exists
func (i *Index) Put(path string, size int, hash string) error {
	_, err := i.db.Exec(put_sql, path, size, hash)
	return err
}

func (i *Index) GetPath(path string) (IndexRow, error) {
	var row IndexRow
	raw := i.db.QueryRow(select_path, path)
	if raw == nil {
		return row, ErrResourceNotFound
	}

	err := raw.Scan(&row.path, &row.size, &row.hash)
	return row, err
}

func (i *Index) DeletePath(path string) error {
	_, err := i.db.Exec(delete_path, path)
	return err
}
