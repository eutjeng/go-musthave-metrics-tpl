package dbstorage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/jmoiron/sqlx"
)

// DBStorage struct for database storage
type DBStorage struct {
	db *sqlx.DB
}

// Interface is an interface defining the methods that should be
// implemented by the DBStorage struct for database storage
type Interface interface {
	models.GeneralStorageInterface          // Embedding a general storage interface
	Ping() error                            // Method to ping the database to ensure the connection
	CreateTables(ctx context.Context) error // Method to create necessary tables
	Close() error                           // Method to close the database connection
}

// NewDBStorage initializes and returns a new DBStorage object
func NewDBStorage(cfg *config.Config) (*DBStorage, error) {

	db, err := sqlx.Open("postgres", cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Second)

	storage := &DBStorage{db: db}

	return storage, nil
}

// Close is responsible for closing the database connection
func (s *DBStorage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}

	return nil
}

// Ping checks if the database is reachable
func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

// CreateTables is responsible for creating the necessary tables in the database
func (s *DBStorage) CreateTables(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gauges (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			value DOUBLE PRECISION NOT NULL
		);
		CREATE TABLE IF NOT EXISTS counters (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			value BIGINT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS gauges_id_index ON gauges (id)")
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS counters_id_index ON counters (id)")
	if err != nil {
		return err
	}

	return nil
}

// UpdateGauge updates a gauge metric in the database
func (s *DBStorage) UpdateGauge(ctx context.Context, name string, value float64, shouldNotify bool) error {
	stmt, err := s.db.PreparexContext(ctx, `
			INSERT INTO gauges (name, value) VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, value)
	return err
}

// UpdateCounter updates a counter metric in the database
func (s *DBStorage) UpdateCounter(ctx context.Context, name string, value int64, shouldNotify bool) error {
	stmt, err := s.db.PreparexContext(ctx, `
			INSERT INTO counters (name, value) VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, value)
	return err
}

// GetGauge retrieves a gauge metric by its name from the database
func (s *DBStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	err := s.db.QueryRowContext(ctx, "SELECT value FROM gauges WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// GetCounter retrieves a counter metric by its name from the database
func (s *DBStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	var value int64
	err := s.db.QueryRowContext(ctx, "SELECT value FROM counters WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// SaveMetrics saves an array of metrics into the database in a single transaction
func (s *DBStorage) SaveMetrics(ctx context.Context, metrics []models.Metrics, shouldNotify bool) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
			}
		} else {
			if err = tx.Commit(); err != nil {
				err = fmt.Errorf("commit failed: %v", err)
			}
		}
	}()

	gaugeStmt, err := tx.PrepareNamedContext(ctx, `
	INSERT INTO gauges (name, value) VALUES (:name, :value)
	ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;
`)
	if err != nil {
		return err
	}
	defer gaugeStmt.Close()

	counterStmt, err := tx.PrepareNamedContext(ctx, `
	INSERT INTO counters (name, value) VALUES (:name, :value)
	ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value;
`)
	if err != nil {
		return err
	}
	defer counterStmt.Close()

	for _, metric := range metrics {
		args := map[string]interface{}{
			"name":  metric.ID,
			"value": nil,
		}

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return fmt.Errorf("value not provided for gauge: %s", metric.ID)
			}
			args["value"] = *metric.Value
			_, err = gaugeStmt.ExecContext(ctx, args)
			if err != nil {
				return err
			}
		case "counter":
			if metric.Delta == nil {
				return fmt.Errorf("delta not provided for counter: %s", metric.ID)
			}
			args["value"] = *metric.Delta
			_, err = counterStmt.ExecContext(ctx, args)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	return tx.Commit()
}

// String fetches and formats metrics for a string representation
func (s *DBStorage) String(ctx context.Context) string {
	var result strings.Builder

	result.Grow(1024)

	if err := s.fetchAndFormat(ctx, "SELECT name, value FROM gauges", "Gauge values:\n", &result, true); err != nil {
		result.WriteString(fmt.Sprintf("Error fetching gauges: %s\n", err.Error()))
	}
	result.WriteString("\n")
	if err := s.fetchAndFormat(ctx, "SELECT name, value FROM counters", "Counter values:\n", &result, false); err != nil {
		result.WriteString(fmt.Sprintf("Error fetching counters: %s\n", err.Error()))
	}

	return result.String()
}

// fetchAndFormat fetches metrics from the database and writes them into a string builder
func (s *DBStorage) fetchAndFormat(ctx context.Context, query, header string, builder io.StringWriter, isFloat bool) error {
	rows, err := s.db.QueryContext(ctx, query)
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
			var value int64
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
