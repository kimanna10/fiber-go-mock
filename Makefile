include .env
export

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Volumes -v
# Detached mode -d (запуск на фоне)

up:
	docker compose up -d

# Остановка и удаление контейнеров
down:
	docker compose down 

# Просмотр запущенных контейнеров
ps:
	docker compose ps

# Полный перезапуск с пересборкой
reset:
	docker compose down -v
	docker compose up -d --build

# Остановка и очистка volumes
clean:
	docker compose down -v

# Управление отдельными базами
postgres-up:
	docker compose up -d fiber-postgres

postgres-stop:
	docker compose stop fiber-postgres

redis-up:
	docker compose up -d fiber-redis

redis-stop:
	docker compose stop fiber-redis

# Создание миграции: make migrate-create seq=users
migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует название файла! Пример: make migrate-create seq=init" \
		exit 1; \
	fi;
	migrate create -ext sql -dir migrations -seq "$(seq)"

# Накат миграций
migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

# Откат миграций
migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down


V ?= 0

migrate-force:
	migrate -path ./migrations -database "$(DB_URL)" force $(V)


# Запуск приложения
run:
	go run ./cmd/server/main.go

worker-run:
	go run ./cmd/worker/main.go
