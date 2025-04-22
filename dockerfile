FROM golang:1.23

ARG APP_DIR
ARG APP_SRC
ENV APP_DIR=${APP_DIR}
ENV APP_SRC=${APP_SRC}

# Install tools
RUN apt-get update && \
    apt-get install -y vim

# Copy source
COPY . ${APP_DIR}

# Install go modules
WORKDIR ${APP_SRC}
RUN go mod download

CMD ["tail", "-f", "/dev/null"]