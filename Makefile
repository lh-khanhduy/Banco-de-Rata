DB_URL=postgresql://root:khanhduy@localhost:5432/banco_de_rata?sslmode=disable
	
postgres:
	docker run --name postgres18 \
		--network bank-network \
		-p 5432:5432 \
		-e POSTGRES_USER=root \
		-e POSTGRES_PASSWORD=khanhduy \
		-d \
		postgres:18-alpine

createdb:
	docker exec -it postgres18 createdb --username=root --owner=root banco_de_rata 
	
dropdb:
	docker exec -it postgres18 dropdb banco_de_rata 
	
migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up
	
migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1
	
migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down
	
migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate
	
test:
	go test -v -cover ./...
	
server:
	go run main.go
	
.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server