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
	
new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate
	
test:
	go test -v -cover ./...
	
server:
	go run main.go

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=banco_de_rata \
    proto/*.proto
	
evans:
	evans --host localhost --port 6767 -r repl

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 db_docs db_schema sqlc test server proto evans new_migration