package database

import (
	"context"
	"errors"
	"gameserver/internal/database/sql"
	"gameserver/internal/interfaces"
	"github.com/rs/zerolog/log"
	"os"
)

// Factory creates and initializes database services
type Factory struct {
	env interfaces.Environment
}

// NewDatabaseFactory creates a new database factory
func NewDatabaseFactory(env interfaces.Environment) *Factory {
	return &Factory{
		env: env,
	}
}

// CreateDatabaseService creates and initializes a new database service
func (f *Factory) CreateDatabaseService(ctx context.Context) (*DatabaseService, error) {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default to SQLite in development
		if f.env == interfaces.Development {
			dbURL = "file:./db.sqlite?cache=shared&mode=rwc"
		} else {
			return nil, errors.New("DATABASE_URL environment variable not set")
		}
	}

	// Create SQL client
	opts := sql.WithAllowedTables([]string{"stories"})
	client, err := sql.NewClient(ctx, dbURL, opts)
	if err != nil {
		return nil, err
	}

	// Initialize database schema
	if err := f.initializeSchema(ctx, client); err != nil {
		client.Close()
		return nil, err
	}

	log.Info().Str("driver", client.Driver()).Msg("SQL database client initialized for tell-it")
	return NewDatabaseService(client), nil
}

// initializeSchema creates the necessary tables if they don't exist
func (f *Factory) initializeSchema(ctx context.Context, client *sql.Client) error {
	// Create stories table
	var createStoriesTable string
	if client.Driver() == "postgres" {
		createStoriesTable = `
			CREATE TABLE IF NOT EXISTS stories (
				id SERIAL PRIMARY KEY,
				text TEXT NOT NULL,
				author TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`
	} else {
		createStoriesTable = `
			CREATE TABLE IF NOT EXISTS stories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				text TEXT NOT NULL,
				author TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`
	}

	if err := client.Exec(ctx, createStoriesTable); err != nil {
		return err
	}

	return nil
}
