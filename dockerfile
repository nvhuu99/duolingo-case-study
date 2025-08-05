FROM golang:1.23.3

# Build args
ARG SVC_DIR_NAME

# Environment vars
ENV SRC_DIR=/var/duolingo/src
ENV SVC_DIR=${SRC_DIR}/apps/${SVC_DIR_NAME}

# Dev tools (optional)
RUN apt-get update && \
    apt-get install -y vim && \
    rm -rf /var/lib/apt/lists/*

# Copy go.mod and go.sum first to leverage caching for downloading dependencies
WORKDIR ${SRC_DIR}
COPY ./src/go.mod ./
COPY ./src/go.sum ./
RUN go mod download

# Copy src & Build
COPY ./src/ ./
WORKDIR ${SVC_DIR}
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/server ./main/main.go && \
    chmod +x ./bin/server

ENTRYPOINT ["./bin/server"]
