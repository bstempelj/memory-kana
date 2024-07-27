FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod ./
#COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o memory-kana

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app /

EXPOSE 1234

USER nonroot:nonroot

ENTRYPOINT ["./memory-kana"]
