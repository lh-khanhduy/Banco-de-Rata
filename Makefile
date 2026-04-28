DB_URL=postgresql://root:khanhduy@localhost:5432/banco_de_rata?sslmode=disable
	
postgres:
	docker run --name postgres18 \
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
	
migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate
	
.PHONY: postgres createdb dropdb migrateup migratedown sqlc