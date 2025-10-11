# Build stage
FROM golang:1.25 AS build

WORKDIR /go/src/memory-kana

COPY go.mod go.sum .
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/memory-kana

# Release stage
FROM gcr.io/distroless/static-debian12
LABEL org.opencontainers.image.source=https://github.com/bstempelj/memory-kana

COPY --from=build /go/bin/memory-kana /
EXPOSE 1234
USER nonroot:nonroot
ENTRYPOINT ["/memory-kana"]
