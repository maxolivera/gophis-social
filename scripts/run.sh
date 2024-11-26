
go fmt ./...
swag init -g app.go -d internal/api,internal/storage/models && swag fmt

go run ./cmd/server
