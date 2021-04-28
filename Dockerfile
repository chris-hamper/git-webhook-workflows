FROM golang:1.16 as build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o git-webhook-workflows

# Generate a minimal image
FROM gcr.io/distroless/base

COPY --from=build /app/git-webhook-workflows /usr/local/bin/

EXPOSE 5000
USER 1000

ENTRYPOINT [ "git-webhook-workflows" ]
