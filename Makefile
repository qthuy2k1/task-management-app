postgres:
	docker exec -it task-management-app-postgres-1 bash

migrate:
	docker run --rm -it -v C:/Users/thuy.nguyen/go/task-management-app/db/migrations:/migrations migrate/migrate -path=/migrations/ --network task-management-app_new -database "postgres://postgres:root@postgres:5432/task-management-db?sslmode=disable" up

sqlboiler:
	docker run --rm -it -v C:/Users/thuy.nguyen/go/task-management-app/sqlboiler.toml:/sqlboiler.toml:ro -v C:/Users/thuy.nguyen/go/task-management-app/models:/models:rw --network task-management-app_new goodwithtech/sqlboiler:latest --wipe /sqlboiler-psql --output models/gen

.PHONY: postgres migrate