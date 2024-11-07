FROM golang:1.23.2 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o memory-kana

FROM gcr.io/distroless/base-debian12 AS release-stage

WORKDIR /

COPY --from=build-stage /app /

EXPOSE 1234

USER nonroot:nonroot

ENTRYPOINT ["./memory-kana"]
