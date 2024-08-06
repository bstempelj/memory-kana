# Memory Kana

## How to run?

Install

- `docker`
- `docker compose`
- `direnv`
- `make`
- `go`

Create an `.envrc` file by copying `.envrc.example`

```sh
cp .envrc.example .env
```

and allow it's execution

```sh
direnv allow
```

### Development build

Install `postgres` build downloading its docker image with

```sh
make dev
```

build the app and run it

```sh
go build && ./memory-kana
# or
go run .
```

then open your browser and go to `localhost:1234`.

### Production build

Just run

```sh
make prod
```

then open your browser and go to `localhost:1234`.

## How it looks
![](howitlooks.png)
