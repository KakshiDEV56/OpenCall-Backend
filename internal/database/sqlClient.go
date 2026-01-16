package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Driver            string
	Host              string
	Port              int
	User              string
	Password          string
	DBName            string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxIdleTime   time.Duration
	ConnectionTimeout time.Duration
}

type SQLClient struct {
	DB *sql.DB
}

func NewSQLClient(cfg Config) (*SQLClient, error) {
	dbSource := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
	)

	db, err := sql.Open(cfg.Driver, dbSource)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectionTimeout)
	defer cancel()

	// Ping the database to ensure a successful connection
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &SQLClient{DB: db}, nil
}

func (c *SQLClient) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
