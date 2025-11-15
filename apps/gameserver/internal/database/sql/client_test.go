package sql

import (
	"context"
	"os"
	"testing"
)

type TestUser struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func setupTestDB(t *testing.T) Database {
	// Create temp database file
	dbFile := "test_db.sqlite"

	// Clean up any existing test database
	os.Remove(dbFile)

	ctx := context.Background()
	db, err := New(ctx, dbFile, WithAllowedTables([]string{"test_users"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create test table
	createTableSQL := `
		CREATE TABLE test_users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		)
	`

	err = db.Exec(ctx, createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func teardownTestDB(db Database) {
	db.Close()
	os.Remove("test_db.sqlite")
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()
	user := TestUser{
		ID:    "user1",
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Log("✓ Create operation successful")
}

func TestGet(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()
	user := TestUser{
		ID:    "user2",
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	// Create user first
	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Get user
	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user2", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != "Jane Smith" {
		t.Errorf("Expected name 'Jane Smith', got '%s'", retrieved.Name)
	}

	t.Log("✓ Get operation successful")
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()
	user := TestUser{
		ID:    "user3",
		Name:  "Bob Jones",
		Email: "bob@example.com",
	}

	// Create user
	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update user
	user.Email = "bob.jones@example.com"
	err = db.Update(ctx, "test_users", "user3", user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user3", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Email != "bob.jones@example.com" {
		t.Errorf("Expected email 'bob.jones@example.com', got '%s'", retrieved.Email)
	}

	t.Log("✓ Update operation successful")
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()
	user := TestUser{
		ID:    "user4",
		Name:  "Alice Brown",
		Email: "alice@example.com",
	}

	// Create user
	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete user
	err = db.Delete(ctx, "test_users", "user4")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user4", &retrieved)
	if err == nil {
		t.Error("Expected error when getting deleted user, got nil")
	}

	t.Log("✓ Delete operation successful")
}

func TestQuery(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// Create multiple users
	users := []TestUser{
		{ID: "user5", Name: "User Five", Email: "five@example.com"},
		{ID: "user6", Name: "User Six", Email: "six@example.com"},
		{ID: "user7", Name: "User Seven", Email: "seven@example.com"},
	}

	for _, user := range users {
		err := db.Create(ctx, "test_users", user)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// Query all users
	var results []TestUser
	err := db.Query(ctx, "SELECT * FROM test_users ORDER BY id", &results)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 users, got %d", len(results))
	}

	t.Log("✓ Query operation successful")
}

func TestTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	// Create user in transaction
	user := TestUser{
		ID:    "user8",
		Name:  "Transaction User",
		Email: "tx@example.com",
	}

	err = tx.Create(ctx, "test_users", user)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Transaction Create failed: %v", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify user exists
	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user8", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	t.Log("✓ Transaction commit successful")
}

func TestTransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	// Create user in transaction
	user := TestUser{
		ID:    "user9",
		Name:  "Rollback User",
		Email: "rollback@example.com",
	}

	err = tx.Create(ctx, "test_users", user)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Transaction Create failed: %v", err)
	}

	// Rollback transaction
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify user does not exist
	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user9", &retrieved)
	if err == nil {
		t.Error("Expected error when getting rolled back user, got nil")
	}

	t.Log("✓ Transaction rollback successful")
}

func TestSQLInjectionInUsername(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// Attempt SQL injection in username field
	user := TestUser{
		ID:    "user10",
		Name:  "John'; DROP TABLE test_users; --",
		Email: "john@example.com",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create with injection attempt failed: %v", err)
	}

	// Verify table still exists and data was stored safely
	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user10", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != "John'; DROP TABLE test_users; --" {
		t.Errorf("Expected name with SQL, got '%s'", retrieved.Name)
	}

	t.Log("✓ SQL injection in username safely escaped")
}

func TestSQLInjectionWithQuotes(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	user := TestUser{
		ID:    "user11",
		Name:  "'; DELETE FROM test_users WHERE '1'='1",
		Email: "test@example.com",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify no data was deleted
	var results []TestUser
	err = db.Query(ctx, "SELECT * FROM test_users", &results)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 user (injection prevented), got %d", len(results))
	}

	t.Log("✓ SQL injection with quotes safely handled")
}

func TestSQLInjectionInEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	user := TestUser{
		ID:    "user12",
		Name:  "Normal Name",
		Email: "test@example.com' OR '1'='1",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user12", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Email != "test@example.com' OR '1'='1" {
		t.Errorf("Expected email with injection payload, got '%s'", retrieved.Email)
	}

	t.Log("✓ SQL injection in email safely escaped")
}

func TestSQLInjectionInID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	user := TestUser{
		ID:    "user13' OR '1'='1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user13' OR '1'='1", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != "Test User" {
		t.Errorf("Expected to retrieve correct user by injection ID, got '%s'", retrieved.Name)
	}

	t.Log("✓ SQL injection in ID safely handled")
}

func TestSQLInjectionWithBackslashes(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	user := TestUser{
		ID:    "user14",
		Name:  "Test\\'; DROP TABLE test_users; --",
		Email: "test@example.com",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user14", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != "Test\\'; DROP TABLE test_users; --" {
		t.Errorf("Expected name with backslashes, got '%s'", retrieved.Name)
	}

	t.Log("✓ SQL injection with backslashes safely escaped")
}

func TestSQLInjectionUnionBased(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	user := TestUser{
		ID:    "user15",
		Name:  "User Name",
		Email: "test@example.com' UNION SELECT * FROM test_users --",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user15", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Email != "test@example.com' UNION SELECT * FROM test_users --" {
		t.Errorf("Expected email with UNION injection, got '%s'", retrieved.Email)
	}

	t.Log("✓ UNION-based SQL injection safely escaped")
}

func TestSQLInjectionDoubleQuotes(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	user := TestUser{
		ID:    "user16",
		Name:  `Test" OR "1"="1`,
		Email: "test@example.com",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user16", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != `Test" OR "1"="1` {
		t.Errorf("Expected name with double quotes, got '%s'", retrieved.Name)
	}

	t.Log("✓ SQL injection with double quotes safely escaped")
}

func TestSQLInjectionDoubleQuotesDrop(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	user := TestUser{
		ID:    "user17",
		Name:  `Test"; DROP TABLE test_users; --`,
		Email: "test@example.com",
	}

	err := db.Create(ctx, "test_users", user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var retrieved TestUser
	err = db.Get(ctx, "test_users", "user17", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != `Test"; DROP TABLE test_users; --` {
		t.Errorf("Expected name with double quotes, got '%s'", retrieved.Name)
	}

	t.Log("✓ SQL injection with double quotes safely escaped")
}
