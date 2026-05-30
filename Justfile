# List all recipes
_default:
	@just --list --unsorted --color=auto

# Run go fmt recursively
fmt:
	go fmt ./...

# Run go test with verbose flag recursively
test:
	go test -v ./...

# Build the memory-kana docker image
docker-build:
	docker build -t memory-kana .

# Tag the built memory-kana docker image
docker-tag tag:
	docker tag memory-kana blazstempelj/memory-kana:{{tag}}

# Push the tagged memory-kana docker image to docker hub
docker-push tag:
	docker image push blazstempelj/memory-kana:{{tag}}

# Run compose with the profile set with COMPOSE_PROFILES env variable
compose:
	docker compose up -d

# Run compose down with the profile set with COMPOSE_PROFILES env variable
compose-down:
	docker compose down

# Exec into postgres database
db-exec:
	docker compose exec -u postgres postgres \
	psql --username "$PGUSER" --dbname "$PGDATABASE"

# Dump database data (in pg_restore format)
db-dump:
	docker compose exec -u postgres postgres \
	pg_dump --data-only --format=custom --username="$PGUSER" --dbname="$PGDATABASE" \
	| tee "dbdump_$(date +%Y%m%d%H%M%S)" >/dev/null

# Restore database data from a pg_restore formatted dump
db-restore file:
	docker compose exec -T -u postgres postgres \
	pg_restore --data-only --username "$PGUSER" --dbname "$PGDATABASE" < {{file}}

# Remove database volume
clean: compose-down
	docker volume rm memory-kana_pg_data
