.PHONY: postgres createdb dropdb migrateup migratedown sqlc test

# Start PostgreSQL container
postgres:
	docker run --name postgres -p 5432:5432 \
	-e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password \
	-d postgres:17-alpine

# Create the database
createdb:
	docker exec -it postgres createdb --username=postgres --owner=postgres simple_bank

# Drop the database
dropdb:
	docker exec -it postgres dropdb --username=postgres simple_bank

# Run database migrations up
migrateup: createdb  # Ensure the database is created before migrating
	migrate -path db/migration \
	-database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" \
	-verbose up

# Run database migrations down
migratedown:
	migrate -path db/migration \
	-database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" \
	-verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...
