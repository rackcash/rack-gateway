FROM golang:1.22-alpine AS dev

WORKDIR /root/app

ARG BLOCKCHAIN_PATH=blockchain
ARG BLOCKCHAIN_BIN=blockchain_bin
ARG PKG_PATH=pkg

COPY ${BLOCKCHAIN_PATH}/ ./${BLOCKCHAIN_PATH}/
COPY ${PKG_PATH}/ ./${PKG_PATH}/
RUN rm -r ${PKG_PATH}/protos

RUN go mod init infra && go mod tidy
RUN cd ${BLOCKCHAIN_PATH} && go build -o ${BLOCKCHAIN_BIN} && cd - && mv ${BLOCKCHAIN_PATH}/${BLOCKCHAIN_BIN} ./

CMD ["/bin/sh", "-c", "./blockchain_bin"]
