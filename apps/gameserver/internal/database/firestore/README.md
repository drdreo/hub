# Firestore Game Database Package

This package provides a Firestore client for data persistence.

## Setup

1. Create a Firebase project and enable Firestore
2. Download your service account credentials JSON file
3. Place the credentials file in `internal/database/firestore/credentials/service-account.json`
4. Add your project ID to environment variables or pass it directly when creating the client

## Usage

```go
// Create a new client
client, err := firestore.NewClient(ctx,
    firestore.WithProjectID("your-game-project-id"),
    firestore.WithCredentialsFile("path/to/credentials.json"))
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
defer client.Close()
```
