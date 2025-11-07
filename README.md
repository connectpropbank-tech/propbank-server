# Go Backend Setup for ShoPROP

## Requirements
- Go 1.21+

## Setup
1. Update the connection string in `main.go` with your PostgreSQL credentials.
2. Run the backend:

```sh
go run main.go
```

## Dependencies
- github.com/lib/pq (PostgreSQL driver)

## Next Steps
- Add environment variable support for DB credentials
- Implement REST API endpoints
