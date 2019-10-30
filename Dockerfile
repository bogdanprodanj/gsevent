FROM golang:1.12.5 as base
WORKDIR /app

# Run the build
FROM base AS build
ADD . /app
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o ./bin/gsevent ./cmd/aggregator/...

# Build the target runtime layer
FROM alpine:3.10.2 as runtime
COPY --from=build /app/bin/gsevent /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gsevent", "serve"]
