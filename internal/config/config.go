package config

import (
	"fmt"
	"os"
	"regexp"
)

// safeIdentRe validates DB_MOCKS_SCHEMA against SQL injection via identifier names.
var safeIdentRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type Config struct {
	Port          string
	DBName        string
	DBPort        string
	DBWriterHost  string
	DBReaderHost  string
	DBSSLMode     string
	DBUser        string
	DBSchema      string
	DBPassword    string
	PublicBaseURL string
}

func Load() (*Config, error) {
	schema := getenv("DB_MOCKS_SCHEMA", "public")
	if !safeIdentRe.MatchString(schema) {
		return nil, fmt.Errorf(
			"DB_MOCKS_SCHEMA %q is not a valid SQL identifier (letters, digits, underscore; must start with letter or underscore)",
			schema,
		)
	}

	cfg := &Config{
		Port:          getenv("PORT", "8080"),
		DBPort:        getenv("DB_PORT", "5432"),
		DBSSLMode:     getenv("DB_SSL_MODE", "disable"),
		DBSchema:      schema,
		PublicBaseURL: os.Getenv("PUBLIC_BASE_URL"),
	}

	var missing []string
	cfg.DBName = required("DB_NAME", &missing)
	cfg.DBWriterHost = required("DB_WRITER_HOST", &missing)
	cfg.DBReaderHost = required("DB_READER_HOST", &missing)
	cfg.DBUser = required("DB_MOCKS_USER", &missing)
	cfg.DBPassword = required("DB_MOCKS_PASSWORD", &missing)

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func required(key string, missing *[]string) string {
	v := os.Getenv(key)
	if v == "" {
		*missing = append(*missing, key)
	}
	return v
}
