FROM golang:1.22-alpine AS dev

WORKDIR /root/app

ARG LOGGER_PATH=logger
ARG PKG_PATH=pkg
ARG LOGGER_BIN=logger_bin

ARG WEB_PATH=api
ARG WEB_BIN=web

COPY ${WEB_PATH}/ ./${WEB_PATH}/
COPY ${LOGGER_PATH}/ ./${LOGGER_PATH}/
COPY ${PKG_PATH}/ ./${PKG_PATH}/
COPY .env .

RUN go mod init infra
RUN go mod tidy

RUN cd ${LOGGER_PATH} && go build -o ${LOGGER_BIN} && cd - && mv ${LOGGER_PATH}/${LOGGER_BIN} ./

RUN cd ${WEB_PATH} && go build -o ${WEB_BIN} && cd - && mv ${WEB_PATH}/${WEB_BIN} ./

CMD ["/bin/sh", "-c", "./logger_bin & ./web"]
