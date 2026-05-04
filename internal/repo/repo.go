package repo

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	// TODO: replace SEU_USUARIO with your GitHub username
	"github.com/SEU_USUARIO/mock-route-factory/internal/model"
)

var ErrNotFound = errors.New("mock not found")

type Repo struct {
	writer *sql.DB
	reader *sql.DB
	table  string
}

func New(writer, reader *sql.DB, schema string) *Repo {
	return &Repo{
		writer: writer,
		reader: reader,
		table:  pq.QuoteIdentifier(schema) + "." + pq.QuoteIdentifier("http_route_mocks"),
	}
}

func (r *Repo) Upsert(method, path string, status int, body []byte) (*model.Mock, error) {
	q := fmt.Sprintf(`
		INSERT INTO %s (method, path, status, response_body, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (method, path) DO UPDATE SET
			status        = EXCLUDED.status,
			response_body = EXCLUDED.response_body,
			updated_at    = NOW()
		RETURNING id, method, path, status, response_body, updated_at
	`, r.table)

	var m model.Mock
	err := r.writer.QueryRow(q, method, path, status, body).
		Scan(&m.ID, &m.Method, &m.Path, &m.Status, &m.ResponseBody, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repo) Get(method, path string) (*model.Mock, error) {
	q := fmt.Sprintf(`
		SELECT id, method, path, status, response_body, updated_at
		FROM %s
		WHERE method = $1 AND path = $2
	`, r.table)

	var m model.Mock
	err := r.reader.QueryRow(q, method, path).
		Scan(&m.ID, &m.Method, &m.Path, &m.Status, &m.ResponseBody, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repo) List() ([]*model.Mock, error) {
	q := fmt.Sprintf(`
		SELECT id, method, path, status, response_body, updated_at
		FROM %s
		ORDER BY method, path
	`, r.table)

	rows, err := r.reader.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mocks []*model.Mock
	for rows.Next() {
		var m model.Mock
		if err := rows.Scan(&m.ID, &m.Method, &m.Path, &m.Status, &m.ResponseBody, &m.UpdatedAt); err != nil {
			return nil, err
		}
		mocks = append(mocks, &m)
	}
	return mocks, rows.Err()
}

func (r *Repo) Delete(method, path string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE method = $1 AND path = $2`, r.table)
	res, err := r.writer.Exec(q, method, path)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
