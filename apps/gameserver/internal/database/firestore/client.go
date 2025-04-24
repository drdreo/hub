package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"os"
	"sync"
	"time"
)

// Client wraps the Firestore client and provides game-specific operations
type Client struct {
	client      *firestore.Client
	projectID   string
	credPath    string
	credentials []byte
	collection  string
	mu          sync.RWMutex
	timeout     time.Duration
}

// ClientOption allows for customization of the Firestore client
type ClientOption func(*Client)

// WithCredentialsFile sets a custom credentials file path
func WithCredentialsFile(path string) ClientOption {
	return func(c *Client) {
		c.credPath = path
	}
}

// WithCredentials sets custom credentials
func WithCredentials(credentials []byte) ClientOption {
	return func(c *Client) {
		c.credentials = credentials
	}
}

// WithProjectID sets a custom project ID
func WithProjectID(projectID string) ClientOption {
	return func(c *Client) {
		c.projectID = projectID
	}
}

// WithCollection sets the default collection name
func WithCollection(collection string) ClientOption {
	return func(c *Client) {
		c.collection = collection
	}
}

// WithTimeout option for client initialization
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// NewClient creates a new Firestore client with the given options
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	client := &Client{
		credPath: "internal/database/firestore/credentials/service-account.json",
		timeout:  30 * time.Second, // Default timeout
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Create context with timeout for connection
	connCtx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	// Validate required fields
	if client.projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	var firestoreClient *firestore.Client
	var err error

	if client.credentials != nil {
		firestoreClient, err = firestore.NewClient(connCtx, client.projectID, option.WithCredentialsJSON(client.credentials))
	} else if _, err = os.Stat(client.credPath); err == nil {
		firestoreClient, err = firestore.NewClient(connCtx, client.projectID, option.WithCredentialsFile(client.credPath))
	} else {
		firestoreClient, err = firestore.NewClient(connCtx, client.projectID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	client.client = firestoreClient
	return client, nil
}

// WithTimeout creates a context with a timeout for operations
func (c *Client) WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// Close closes the Firestore client
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Collection returns a reference to the specified Firestore collection
func (c *Client) Collection(name string) *firestore.CollectionRef {
	return c.client.Collection(name)
}

// FirestoreClient returns the underlying Firestore client
func (c *Client) FirestoreClient() *firestore.Client {
	return c.client
}
