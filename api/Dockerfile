FROM golang:1.22-alpine AS dev

WORKDIR /root/app

ARG LOGGER_PATH=rackLogger
ARG LOGGER_BIN=logger

ARG WEB_PATH=api
ARG WEB_BIN=web

ARG SECRETS=secrets
ENV SECRETS=${SECRETS}

COPY ${WEB_PATH}/ ./${WEB_PATH}/
COPY ${LOGGER_PATH}/ ./${LOGGER_PATH}/
COPY config.toml .
COPY .env .

ENV ENVPATH=./.env
ENV CONFIG=./config.toml

RUN cd ${LOGGER_PATH} && go build -o ${LOGGER_BIN} && cd - && mv ${LOGGER_PATH}/${LOGGER_BIN} ./

RUN cd ${WEB_PATH} && go build -o ${WEB_BIN} && cd - && mv ${WEB_PATH}/${WEB_BIN} ./

RUN mkdir rack && mv api/rack/frontend rack/ && mv api/rack/templates rack/

CMD ["/bin/sh", "-c", "./logger & ./web"]
