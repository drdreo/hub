package firestore

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Create adds a new document to the specified collection with the given ID
// using .Set over .Create (create fails if the document exists. Set, replaces an existing document or creates a new one)
func (c *Client) Create(ctx context.Context, collection string, data interface{}) error {
	ref := c.client.Collection(collection).NewDoc()
	_, err := ref.Set(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to create document %s in collection %s: %w", ref.ID, collection, err)
	}
	return nil
}

// Get retrieves a document by ID from the specified collection
func (c *Client) Get(ctx context.Context, collection string, id string, dest interface{}) error {
	doc, err := c.client.Collection(collection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("document %s not found in collection %s", id, collection)
		}
		return fmt.Errorf("failed to get document %s from collection %s: %w", id, collection, err)
	}

	if err = doc.DataTo(dest); err != nil {
		return fmt.Errorf("failed to parse document data: %w", err)
	}

	return nil
}

// Update updates fields in the document with the given ID
func (c *Client) Update(ctx context.Context, collection string, id string, updates []firestore.Update) error {
	_, err := c.client.Collection(collection).Doc(id).Update(ctx, updates)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("document %s not found in collection %s", id, collection)
		}
		return fmt.Errorf("failed to update document %s in collection %s: %w", id, collection, err)
	}
	return nil
}

// Delete removes a document from the specified collection
func (c *Client) Delete(ctx context.Context, collection string, id string) error {
	_, err := c.client.Collection(collection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete document %s from collection %s: %w", id, collection, err)
	}
	return nil
}

// Query performs a query on the specified collection
func (c *Client) Query(ctx context.Context, collection string, queries []Query, dest interface{}) error {
	q := c.client.Collection(collection).Query

	// Apply filters
	for _, query := range queries {
		if query.Field != "" && query.Op != "" {
			q = q.Where(query.Field, query.Op, query.Value)
		}
		if query.OrderBy != "" {
			q = q.OrderBy(query.OrderBy, query.OrderDir)
		}
		if query.Limit > 0 {
			q = q.Limit(query.Limit)
		}
	}

	iter := q.Documents(ctx)
	defer iter.Stop()

	// Check if the destination is a slice
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("destination must be a pointer to a slice")
	}

	sliceVal := destVal.Elem()
	elemType := sliceVal.Type().Elem()

	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating through query results: %w", err)
		}

		// Create a new element of the slice type
		newElem := reflect.New(elemType).Elem()

		// If it's a struct type, try to use DataTo
		if newElem.Kind() == reflect.Struct {
			newElemAddr := reflect.New(elemType)
			if err := doc.DataTo(newElemAddr.Interface()); err != nil {
				return fmt.Errorf("failed to parse document data: %w", err)
			}
			newElem = newElemAddr.Elem()
		} else {
			// Handle map type
			data := doc.Data()
			newElem.Set(reflect.ValueOf(data))
		}

		sliceVal = reflect.Append(sliceVal, newElem)
	}

	destVal.Elem().Set(sliceVal)
	return nil
}

// RunTransaction executes the given function within a Firestore transaction
func (c *Client) RunTransaction(ctx context.Context, f func(context.Context, *firestore.Transaction) error) error {
	return c.client.RunTransaction(ctx, f)
}

// BulkWrite performs multiple writes efficiently but non-atomically
func (c *Client) BulkWrite(ctx context.Context, operations []BulkOperation) error {
	writer := c.client.BulkWriter(ctx)
	defer writer.End()

	for _, op := range operations {
		switch op.Type {
		case BulkCreate:
			writer.Create(c.client.Collection(op.Collection).Doc(op.ID), op.Data)
		case BulkUpdate:
			writer.Update(c.client.Collection(op.Collection).Doc(op.ID), op.Updates)
		case BulkDelete:
			writer.Delete(c.client.Collection(op.Collection).Doc(op.ID))
		}
	}

	writer.Flush()
	return nil
}
