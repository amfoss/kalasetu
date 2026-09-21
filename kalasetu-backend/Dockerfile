
FROM golang:latest AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /kalasetu

# Testing
FROM build-stage AS test-stage

RUN go test -v ./...

# Deploying into a lean image
FROM alpine:latest AS release-stage

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build-stage /kalasetu .

ENTRYPOINT ["./kalasetu"]