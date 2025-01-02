include .env
export $(shell sed 's/=.*//' .env)

.PHONY: tests docker protoc

ARGS := $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))

# config path
WEB_CONFIG = ../config.toml
CURRENCY_CONFIG = ../config.toml
BOT_CONFIG = ./config.toml
LOGGER_ENV = ../.env
WEB_ENV = ../.env
SECRETS = ../secrets
BOT_SECRETS = ./secrets
LOCALES = ./locales

%:
	@:

run_all: run_logger run_web
# run_all: ts run_web

run_bot:
	# CONFIG=$(BOT_CONFIG) SECRETS=$(BOT_SECRETS) LOCALES=$(LOCALES) tsx rackBot/index.ts
	CONFIG=$(BOT_CONFIG) SECRETS=$(BOT_SECRETS) LOCALES=$(LOCALES) bun rackBot/index.ts
run_currency:
	cd blockchain && SECRETS=$(SECRETS) CONFIG=$(CURRENCY_CONFIG) go run .
run_logger:
	# killall racklog || true
	cd rackLogger && ENVPATH=$(LOGGER_ENV) go run . &
run_web:
	cd api && TZ=UTC+3 SECRETS=$(SECRETS) CONFIG=$(WEB_CONFIG) ENVPATH=$(WEB_ENV) go run .

# parseable:
# 	 docker run -p 8000:8000 \
#                          -v /tmp/parseable/data:/parseable/data \
#                          -v /tmp/parseable/staging:/parseable/staging \
#                          -e P_FS_DIR=/parseable/data \
# 						 -e P_SEND_ANONYMOUS_USAGE_DATA=false \
#                          -e P_STAGING_DIR=/parseable/staging \
# 						 -e P_USERNAME=$(PARSEABLE_USERNAME) \
# 						 -e P_PASSWORD=$(PARSEABLE_PASSWORD) \
#                          containers.parseable.com/parseable/parseable:latest \
#                          parseable local-store

# nats:
# 	# docker pull nats:latest
# 	docker run -p 4222:4222 -ti nats:latest
stress: 
	siege -c 100 -t 1M http://localhost:8888
tests_web:
	cd api && go test -timeout 30s rackpay/rack/nats -v && go test -timeout 30s rackpay/rack/app/v1/helpers -v
tests:
	cd tests && go test *.go -v
ts: 
	cd api/internal/frontend && npm run build
docker:
	python3 scripts/docker.py $(ARGS)
# compose:
# 	bash scripts/compose.sh

protoc:
	cd pkg/protos && protoc --go_out=gen/go --go_opt=paths=source_relative \
   --go-grpc_out=gen/go --go-grpc_opt=paths=source_relative log.proto

default: run_all


