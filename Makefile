test: mock
	golangci-lint run ./...
	go test ./... -v

mock:
	mockgen -source=./internal/infrastructure/postgres/repository.go -destination=./internal/infrastructure/postgres/mock/repository_mock.go
	mockgen -source=./internal/infrastructure/postgres/transaction.go -destination=./internal/infrastructure/postgres/mock/transaction_mock.go
	mockgen -source=./internal/application/reservation.go -destination=./internal/application/mock/mock.go
	mockgen -source=./internal/transport/handler.go -destination=./internal/transport/mock/mock.go

run:
	docker-compose build && docker-compose up