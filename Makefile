test: mock
	golangci-lint run ./...
	go test ./... -v -count=1

mock:
	mockgen -source=./internal/domain/repository.go -destination=./internal/domain/mock/repository_mock.go
	mockgen -source=./internal/application/reservation.go -destination=./internal/application/mock/mock.go
	mockgen -source=./internal/transport/handler.go -destination=./internal/transport/mock/mock.go

run:
	docker-compose build && docker-compose up