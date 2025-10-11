# Build stage
FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o memory-kana

FROM gcr.io/distroless/base-debian12

LABEL org.opencontainers.image.source=https://github.com/bstempelj/memory-kana

WORKDIR /

COPY --from=build /app /

EXPOSE 1234

USER nonroot:nonroot

ENTRYPOINT ["./memory-kana"]
