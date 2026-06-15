package db

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

func Connect(ctx context.Context, databaseURL string) (*sql.DB, error) {
	conn, err := sql.Open("sqlserver", databaseURL)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(30 * time.Minute)
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func Migrate(ctx context.Context, conn *sql.DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, batch := range splitBatches(string(raw)) {
		if strings.TrimSpace(batch) == "" {
			continue
		}
		if _, err := conn.ExecContext(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

func splitBatches(script string) []string {
	re := regexp.MustCompile(`(?im)^\s*GO\s*$`)
	return re.Split(script, -1)
}
