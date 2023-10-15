package dbstorage

import (
	"database/sql"
)

// DBStorage struct for database storage
type DBStorage struct {
	DB *sql.DB
}

// StorageInterface defines methods for database storage
type StorageInterface interface {
	Ping() error
}

// NewDBStorage initializes new database storage
func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &DBStorage{DB: db}, nil
}

// Ping checks the database connection
func (s *DBStorage) Ping() error {
	return s.DB.Ping()
}
