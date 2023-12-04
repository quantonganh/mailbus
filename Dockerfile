# Build the application from source
FROM golang:1.21-alpine3.18 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOARCH=arm64 go build -o /mailbus -v -ldflags="-s -w" cmd/mailbus/main.go

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /mailbus /mailbus

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/mailbus"]