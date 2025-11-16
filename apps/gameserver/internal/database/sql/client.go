package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"  // PostgreSQL driver
	_ "modernc.org/sqlite" // SQLite driver
)

// Common errors
var (
	ErrRecordNotFound   = errors.New("record not found")
	ErrInvalidTableName = errors.New("invalid table name")
)

// validTableNames is a whitelist of allowed table names for security
var validTableNames = make(map[string]bool)

type Client struct {
	db              *sqlx.DB
	validTableNames map[string]bool
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
}

// Tx wraps a database transaction
type transaction struct {
	tx     *sqlx.Tx
	db     *sqlx.DB
	client *Client
}

// Compile-time interface assertions
var (
	_ Database    = (*Client)(nil)
	_ Transaction = (*transaction)(nil)
)

// ClientOption allows for customization of the SQL client
type ClientOption func(*Client)

// WithAllowedTables sets custom allowed table names (useful for testing or dynamic tables)
func WithAllowedTables(tables []string) ClientOption {
	return func(c *Client) {
		for _, table := range tables {
			validTableNames[table] = true
		}
	}
}

// WithMaxOpenConns sets the maximum number of open connections to the database
func WithMaxOpenConns(n int) ClientOption {
	return func(c *Client) {
		c.maxOpenConns = n
	}
}

// WithMaxIdleConns sets the maximum number of idle connections in the pool
func WithMaxIdleConns(n int) ClientOption {
	return func(c *Client) {
		c.maxIdleConns = n
	}
}

// WithConnMaxLifetime sets the maximum amount of time a connection may be reused
func WithConnMaxLifetime(d time.Duration) ClientOption {
	return func(c *Client) {
		c.connMaxLifetime = d
	}
}

// WithConnMaxIdleTime sets the maximum amount of time a connection may be idle
func WithConnMaxIdleTime(d time.Duration) ClientOption {
	return func(c *Client) {
		c.connMaxIdleTime = d
	}
}

// validateTableName checks if the table name is in the whitelist
func validateTableName(table string) error {
	if !validTableNames[table] {
		return fmt.Errorf("%w: %s", ErrInvalidTableName, table)
	}
	return nil
}

// extractDBColumns extracts column names from struct db tags or map keys
func extractDBColumns(data interface{}) []string {
	var columns []string

	// 1. Handle map (fast path)
	if dataMap, ok := data.(map[string]interface{}); ok {
		for key := range dataMap {
			columns = append(columns, key)
		}
		return columns
	}

	// 2. Handle struct/slice with reflection
	val := reflect.ValueOf(data)
	// De-reference pointers (e.g., *User or *[]User)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 3. Check for slice or array
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		// Get the *type* of the elements in the slice (e.g., User or *User)
		elemType := val.Type().Elem()

		// If elements are pointers (e.g., []*User), get the base type
		for elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		// Create a new zero-value instance of the element type.
		// This is the key: it's safe for empty slices.
		val = reflect.New(elemType).Elem()
	}

	// 4. At this point, val *should* be a struct
	if val.Kind() != reflect.Struct {
		return columns
	}

	// 5. Extract tags from the struct
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		dbTag := field.Tag.Get("db")
		// Skip empty tags and "-"
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Skip nil pointer fields
		fieldVal := val.Field(i)
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			continue
		}

		columns = append(columns, dbTag)
	}

	return columns
}

// newClient creates a new SQL client with the given connection string
// connectionString format:
// - PostgreSQL: "postgres://user:password@host:port/dbname?sslmode=disable"
// - SQLite: "file:path/to/db.sqlite?cache=shared&mode=rwc"
func newClient(ctx context.Context, connectionString string, opts ...ClientOption) (*Client, error) {
	// Determine driver from connection string
	var driver string
	if strings.HasPrefix(connectionString, "postgres://") || strings.HasPrefix(connectionString, "postgresql://") {
		driver = "postgres"
	} else if strings.HasPrefix(connectionString, "file:") || strings.HasSuffix(connectionString, ".sqlite") || strings.HasSuffix(connectionString, ".db") {
		driver = "sqlite"
		// Ensure SQLite connection string is properly formatted
		if !strings.HasPrefix(connectionString, "file:") {
			connectionString = "file:" + connectionString
		}
	} else {
		return nil, fmt.Errorf("unsupported connection string format: %s", connectionString)
	}

	db, err := sqlx.ConnectContext(ctx, driver, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	client := &Client{
		db:              db,
		validTableNames: make(map[string]bool),
		maxOpenConns:    25,
		maxIdleConns:    5,
		connMaxLifetime: 5 * time.Minute,
		connMaxIdleTime: 15 * time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Apply connection pool settings
	db.SetMaxOpenConns(client.maxOpenConns)
	db.SetMaxIdleConns(client.maxIdleConns)
	db.SetConnMaxLifetime(client.connMaxLifetime)
	db.SetConnMaxIdleTime(client.connMaxIdleTime)

	log.Info().Str("driver", driver).Msg("database client created")

	return client, nil
}

// New creates a new SQL database client and returns it as a Database interface.
// Example:
//
//	db, err := sql.New(ctx, "postgres://user:pass@localhost/mydb")
//	if err != nil {
//	    return err
//	}
//	defer db.Close()
//	// db can be passed to functions expecting Database interface
func New(ctx context.Context, connectionString string, opts ...ClientOption) (Database, error) {
	return newClient(ctx, connectionString, opts...)
}

func (c *Client) Create(ctx context.Context, table string, data interface{}) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	columns := extractDBColumns(data)
	if len(columns) == 0 {
		return fmt.Errorf("no db tags found")
	}

	placeholders := make([]string, len(columns))
	for i, col := range columns {
		placeholders[i] = ":" + col
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := c.db.NamedExecContext(ctx, query, data)
	return err
}

// Get retrieves a record by ID from the specified table
// dest should be a pointer to a struct with db tags
func (c *Client) Get(ctx context.Context, table string, id string, dest interface{}) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", table)
	query = c.db.Rebind(query)

	err := c.db.GetContext(ctx, dest, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: id=%s in table %s", ErrRecordNotFound, id, table)
		}
		return fmt.Errorf("failed to get from %s: %w", table, err)
	}
	return nil
}

// Update updates a record in the specified table
// data should be a struct with db tags or a map[string]interface{}
// The struct/map should include the id field
func (c *Client) Update(ctx context.Context, table string, id string, data interface{}) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	columns := extractDBColumns(data)
	if len(columns) == 0 {
		return fmt.Errorf("no db tags found in struct or empty map")
	}

	// Build SET clauses, excluding id
	setClauses := make([]string, 0, len(columns))
	for _, col := range columns {
		if col != "id" {
			setClauses = append(setClauses, col+" = :"+col)
		}
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = :id", table, strings.Join(setClauses, ", "))

	// If data is a map, ensure it has the id field
	if dataMap, ok := data.(map[string]interface{}); ok {
		dataMap["id"] = id
	}

	result, err := c.db.NamedExecContext(ctx, query, data)
	if err != nil {
		return fmt.Errorf("failed to update data in %s: %w", table, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: id=%s in table %s", ErrRecordNotFound, id, table)
	}

	return nil
}

// Delete removes a record from the specified table
func (c *Client) Delete(ctx context.Context, table string, id string) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	query = c.db.Rebind(query)

	result, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record from %s: %w", table, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: id=%s in table %s", ErrRecordNotFound, id, table)
	}

	return nil
}

// Query performs a custom query and scans results into dest
// dest should be a pointer to a slice of structs with db tags
func (c *Client) Query(ctx context.Context, query string, dest interface{}, args ...interface{}) error {
	err := c.db.SelectContext(ctx, dest, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// Exec executes a query without returning any rows
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// BeginTx starts a new database transaction
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	tx, err := c.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &transaction{
		tx:     tx,
		db:     c.db,
		client: c,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// GetDB returns the underlying sqlx.DB for advanced operations
func (c *Client) GetDB() *sqlx.DB {
	return c.db
}

func (c *Client) Driver() string {
	return c.db.DriverName()
}

// Create inserts a new record into the specified table within a transaction
func (t *transaction) Create(ctx context.Context, table string, data interface{}) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	columns := extractDBColumns(data)
	if len(columns) == 0 {
		return fmt.Errorf("no db tags found")
	}

	placeholders := make([]string, len(columns))
	for i, col := range columns {
		placeholders[i] = ":" + col
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := t.tx.NamedExecContext(ctx, query, data)
	return err
}

// Get retrieves a record by ID from the specified table within a transaction
func (t *transaction) Get(ctx context.Context, table string, id string, dest interface{}) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", table)
	query = t.db.Rebind(query)

	err := t.tx.GetContext(ctx, dest, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: id=%s in table %s", ErrRecordNotFound, id, table)
		}
		return fmt.Errorf("failed to get from %s: %w", table, err)
	}
	return nil
}

// Update updates a record in the specified table within a transaction
func (t *transaction) Update(ctx context.Context, table string, id string, data interface{}) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	// Extract column names from struct db tags
	columns := extractDBColumns(data)
	if len(columns) == 0 {
		return fmt.Errorf("no db tags found in struct or empty map")
	}

	// Build SET clauses, excluding id
	setClauses := make([]string, 0, len(columns))
	for _, col := range columns {
		if col != "id" {
			setClauses = append(setClauses, col+" = :"+col)
		}
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = :id", table, strings.Join(setClauses, ", "))

	// If data is a map, ensure it has the id field
	if dataMap, ok := data.(map[string]interface{}); ok {
		dataMap["id"] = id
	}

	result, err := t.tx.NamedExecContext(ctx, query, data)
	if err != nil {
		return fmt.Errorf("failed to update data in %s: %w", table, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: id=%s in table %s", ErrRecordNotFound, id, table)
	}

	return nil
}

// Delete removes a record from the specified table within a transaction
func (t *transaction) Delete(ctx context.Context, table string, id string) error {
	if err := validateTableName(table); err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	query = t.db.Rebind(query)

	result, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record from %s: %w", table, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: id=%s in table %s", ErrRecordNotFound, id, table)
	}

	return nil
}

// Query performs a custom query within a transaction
func (t *transaction) Query(ctx context.Context, query string, dest interface{}, args ...interface{}) error {
	err := t.tx.SelectContext(ctx, dest, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// Exec executes a query without returning any rows within a transaction
func (t *transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// Commit commits the transaction
func (t *transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *transaction) Rollback() error {
	return t.tx.Rollback()
}
