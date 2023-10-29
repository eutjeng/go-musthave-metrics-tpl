package dbstorage

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
)

// DBStorage struct for database storage
type DBStorage struct {
	DB *sql.DB
}

// Interface defines methods for database storage
type Interface interface {
	models.GeneralStorageInterface
	Ping() error
	String() string
	CreateTables() error
}

// NewDBStorage initializes new database storage
func NewDBStorage(ctx context.Context, dsn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	storage := &DBStorage{DB: db}

	return storage, nil
}

// Close closes the database connection
func (s *DBStorage) Close() error {
	if err := s.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}

	return nil
}

// Ping checks the database connection
func (s *DBStorage) Ping() error {
	return s.DB.Ping()
}

// CreateTables creates necessary tables in the database
func (s *DBStorage) CreateTables() error {
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS gauges (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			value DOUBLE PRECISION NOT NULL
		);
		CREATE TABLE IF NOT EXISTS counters (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			value INT NOT NULL
		);
	`)
	return err
}

// UpdateGauge updates the gauge metric in the database
func (s *DBStorage) UpdateGauge(name string, value float64, shouldNotify bool) error {
	_, err := s.DB.Exec(`
		INSERT INTO gauges (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;
	`, name, value)
	return err
}

// UpdateCounter updates the counter metric in the database
func (s *DBStorage) UpdateCounter(name string, value int64, shouldNotify bool) error {
	_, err := s.DB.Exec(`
		INSERT INTO counters (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;
	`, name, value)
	return err
}

// GetGauge retrieves the gauge metric value from the database
func (s *DBStorage) GetGauge(name string) (float64, error) {
	var value float64
	err := s.DB.QueryRow("SELECT value FROM gauges WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// GetCounter retrieves the counter metric value from the database
func (s *DBStorage) GetCounter(name string) (int64, error) {
	var value int64
	err := s.DB.QueryRow("SELECT value FROM counters WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (s *DBStorage) String() string {
	var result strings.Builder

	result.Grow(1024)

	if err := s.fetchAndFormat("SELECT name, value FROM gauges", "Gauge values:\n", &result, true); err != nil {
		result.WriteString(fmt.Sprintf("Error fetching gauges: %s\n", err.Error()))
	}
	result.WriteString("\n")
	if err := s.fetchAndFormat("SELECT name, value FROM counters", "Counter values:\n", &result, false); err != nil {
		result.WriteString(fmt.Sprintf("Error fetching counters: %s\n", err.Error()))
	}

	return result.String()
}

// fetch and format metrics
func (s *DBStorage) fetchAndFormat(query, header string, builder io.StringWriter, isFloat bool) error {
	rows, err := s.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	if _, err := builder.WriteString(header); err != nil {
		return err
	}

	for rows.Next() {
		var name string
		if isFloat {
			var value float64
			if err := rows.Scan(&name, &value); err != nil {
				return err
			}
			if _, err := builder.WriteString(fmt.Sprintf("%s: %f\n", name, value)); err != nil {
				return err
			}
		} else {
			var value int
			if err := rows.Scan(&name, &value); err != nil {
				return err
			}
			if _, err := builder.WriteString(fmt.Sprintf("%s: %d\n", name, value)); err != nil {
				return err
			}
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
