build:
	@cd server && go build

run: build
	@cd server && ./memory-kana
