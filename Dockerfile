FROM golang:1.21-bookworm as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN make build

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:bookworm-slim
RUN set -x && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates netcat-traditional \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/bin/node /app/node

# Run the web service on container startup.
ENTRYPOINT ["/app/node"]