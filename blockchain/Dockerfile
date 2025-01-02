FROM golang:1.22-alpine AS dev

WORKDIR /root/app

ARG CURRENCY_PATH=blockchain
ARG CURRENCY_BIN=currency

ARG SECRETS=secrets
ENV SECRETS=${SECRETS}

COPY ${CURRENCY_PATH}/ ./${CURRENCY_PATH}/
COPY config.toml .
COPY .env .

ENV ENVPATH=./.env
ENV CONFIG=./config.toml

RUN cd ${CURRENCY_PATH} && go build -o ${CURRENCY_BIN} && cd - && mv ${CURRENCY_PATH}/${CURRENCY_BIN} ./

CMD ["/bin/sh", "-c", "./currency"]
