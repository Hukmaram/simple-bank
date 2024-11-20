DB_URL=postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable
.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock

# Start PostgreSQL container
postgres:
	docker run --name postgres17 -p 5432:5432 \
	-e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password \
	-d postgres:17-alpine

# Create the database
createdb:
	docker exec -it postgres17 createdb --username=postgres --owner=postgres simple_bank

# Drop the database
dropdb:
	docker exec -it postgres17 dropdb --username=postgres simple_bank

# Run database migrations up
migrateup:
		migrate -path db/migration -database "$(DB_URL)" -verbose up

# Run database migrations down
migratedown:
		migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

server:
		go run main.go

test:
	go test -v -cover ./...

mock:
	mockgen -package mockdb -destination db/mock/store.go simplebank/db/sqlc Store
