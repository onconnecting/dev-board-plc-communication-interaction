FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-data-service  ./cmd/main.go


# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY templates ./templates
COPY --from=build-stage /docker-data-service /docker-data-service
COPY .env.docker .env

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/docker-data-service"]