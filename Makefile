docker/build:
	@docker build -t memory-kana .

docker/run:
	@docker run -p 1234:1234 --rm --name memory-kana memory-kana
