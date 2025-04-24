package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
)

// Database defines the operations that can be performed on the Firestore database
type Database interface {
	Create(ctx context.Context, collection string, data interface{}) error
	Get(ctx context.Context, collection string, id string, dest interface{}) error
	Update(ctx context.Context, collection string, id string, updates []firestore.Update) error
	Delete(ctx context.Context, collection string, id string) error
	Query(ctx context.Context, collection string, queries []Query, dest interface{}) error
	RunTransaction(ctx context.Context, f func(context.Context, *firestore.Transaction) error) error
	Close() error
}

// Query represents a Firestore query filter
type Query struct {
	Field    string
	Op       string
	Value    interface{}
	OrderBy  string
	OrderDir firestore.Direction
	Limit    int
}

type BulkOperationType int

const (
	BulkCreate BulkOperationType = iota
	BulkUpdate
	BulkDelete
)

type BulkOperation struct {
	Type       BulkOperationType
	Collection string
	ID         string
	Data       interface{}
	Updates    []firestore.Update
}
