
go fmt ./...
swag init -g app.go -d internal/api,internal/models && swag fmt

go run ./cmd/server
