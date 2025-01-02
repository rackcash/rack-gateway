include .env
export $(shell sed 's/=.*//' .env)

default: run

run:
# curl --location --request PUT 'http://127.0.0.1:8000/api/v1/logstream/test'
	cd racklog && ENVPATH=../.env go run . &
	go run .
db_test:
	go test -timeout 30s  rackpay/rack/postgres -v
test:
	go test -timeout 30s  rackpay/rack/api -v