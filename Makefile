docker/build:
	@docker build -t memory-kana .

docker/run:
	@docker run -p 1234:1234 --rm --name memory-kana memory-kana

compose/up:
	@docker compose up -d

compose/down:
	@docker compose down

lint/compose:
	@yamllint docker-compose.yml
