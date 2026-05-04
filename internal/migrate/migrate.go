package migrate

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

func Run(db *sql.DB, schema string) error {
	qs := pq.QuoteIdentifier(schema)
	qt := pq.QuoteIdentifier("http_route_mocks")
	qi := pq.QuoteIdentifier("idx_http_route_mocks_method_path")
	table := qs + "." + qt

	statements := []string{
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, qs),
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id            BIGSERIAL    PRIMARY KEY,
				method        TEXT         NOT NULL,
				path          TEXT         NOT NULL,
				status        INT          NOT NULL DEFAULT 200,
				response_body JSONB        NOT NULL,
				updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
				UNIQUE (method, path)
			)`, table),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (method, path)`, qi, table),
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
