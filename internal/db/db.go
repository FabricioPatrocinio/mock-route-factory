package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"

	// TODO: replace SEU_USUARIO with your GitHub username
	"github.com/SEU_USUARIO/mock-route-factory/internal/config"
)

func Open(cfg *config.Config) (writer, reader *sql.DB, err error) {
	writer, err = openDB(dsn(cfg.DBWriterHost, cfg))
	if err != nil {
		return nil, nil, fmt.Errorf("writer: %w", err)
	}
	reader, err = openDB(dsn(cfg.DBReaderHost, cfg))
	if err != nil {
		writer.Close()
		return nil, nil, fmt.Errorf("reader: %w", err)
	}
	return writer, reader, nil
}

func openDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func dsn(host string, cfg *config.Config) string {
	q := url.Values{}
	q.Set("sslmode", cfg.DBSSLMode)
	// options=-csearch_path=SCHEMA scopes every connection to the configured schema
	q.Set("options", "-csearch_path="+cfg.DBSchema)
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:     host + ":" + cfg.DBPort,
		Path:     "/" + cfg.DBName,
		RawQuery: q.Encode(),
	}
	return u.String()
}
