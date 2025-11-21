package database

import (
	"context"
	"encoding/json"
	"errors"
	"gameserver/internal/database/firestore"
	"gameserver/internal/interfaces"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

// Factory creates and initializes database services
type Factory struct {
	env            interfaces.Environment
	credentialsDir string
}

// NewDatabaseFactory creates a new database factory
func NewDatabaseFactory(env interfaces.Environment, credentialsDir string) *Factory {
	return &Factory{
		env:            env,
		credentialsDir: credentialsDir,
	}
}

// CreateDatabaseService creates and initializes a new database service
func (f *Factory) CreateDatabaseService(ctx context.Context) (*DatabaseService, error) {
	var serviceAccount []byte
	var err error
	projectId := os.Getenv("OWE_DRAHN_FIREBASE_PROJECT_ID")
	if projectId == "" {
		projectId = "owe-drahn"
	}

	if f.env == interfaces.Development {
		// In development, read from file
		credPath := filepath.Join(f.credentialsDir, "service-account.json")
		serviceAccount, err = os.ReadFile(credPath)
		if err != nil {
			return nil, err
		}
	} else {
		// In production, get credentials from environment variable
		credStr := os.Getenv("GCS_CREDENTIALS")
		if credStr == "" {
			return nil, errors.New("GCS_CREDENTIALS environment variable not set")
		}
		serviceAccount = []byte(credStr)
	}

	var jsonObj map[string]interface{}
	if err = json.Unmarshal(serviceAccount, &jsonObj); err != nil {
		return nil, errors.New("invalid service account format")
	}

	client, err := firestore.NewClient(ctx, firestore.WithCredentials(serviceAccount), firestore.WithProjectID(projectId))
	if err != nil {
		return nil, err
	}

	log.Info().Str("project", projectId).Msg("Firestore client initialized")
	return NewDatabaseService(client), nil
}
