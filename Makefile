postgres:
	docker exec -it task-management-app-postgres-1 bash

migrate:
	docker run -v ${GOPATH}/task-management-app/db/migrations:/migrations --network task-management-app_new migrate/migrate -source=file://${GOPATH}/task-management-app/db/migrations -database postgres://postgres:root@localhost:5432/task-management-db?sslmode=disable -verbose up

.PHONY: postgres migrate