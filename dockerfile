FROM golang:1.23

ARG APP_DIR
ARG APP_SRC
ENV APP_DIR=${APP_DIR}
ENV APP_SRC=${APP_SRC}

# Install tools
RUN apt-get update && \
    apt-get install -y vim

# Install Nodejs 18.x
WORKDIR /home/root
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs

# Install Gulp (required for debugging Golang with Delve)
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN npm install --global gulp gulp-cli

# Copy source
COPY . ${APP_DIR}

# Install node modules
WORKDIR ${APP_DIR}
RUN npm install -y

# Install go modules
WORKDIR ${APP_SRC}
RUN go mod download

CMD ["tail", "-f", "/dev/null"]