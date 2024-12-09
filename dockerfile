FROM golang:1.23

ENV APP_SRC=/var/www/html/app

RUN apt-get update && \
    apt-get install -y vim

RUN go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /home/root

RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs

COPY . ${APP_SRC}

WORKDIR ${APP_SRC}

RUN npm install -y
RUN npm install --global gulp gulp-cli
