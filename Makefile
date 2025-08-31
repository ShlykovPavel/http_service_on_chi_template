tidy:
	go mod tidy

run_tests:
	go test -v ./...

add_migration:
	migrate create -ext sql -dir internal/storage/database/migration -seq add_company_id_in_booking_slots_table

run_migration_up:
	migrate -path internal/storage/database/migration -database "postgresql://postgres:mysecretpassword@localhost:5432/auth_service_db?sslmode=disable" -verbose up

run_migration_down:
	migrate -path internal/storage/database/migration -database "postgresql://postgres:mysecretpassword@localhost:5434/booker_service_db?sslmode=disable" -verbose down

force_migration:
	migrate -path internal/storage/database/migration -database "postgresql://postgres:mysecretpassword@localhost:5434/booker_service_db?sslmode=disable" force 7

swagger_generation:
	swag init -g cmd/auth-JWT-microservice/main.go --output docs --parseDependency --parseInternal