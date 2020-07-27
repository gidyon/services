API_IN_PATH := api/proto
API_OUT_PATH := /Users/jessegitaka/go/src
SWAGGER_DOC_OUT_PATH := api/swagger
PROJECT_ROOT := /Users/jessegitaka/go/src/github.com/gidyon/services

setup_dev: ## Sets up a development environment for the emrs project
	@cd deployments/compose/dev &&\
	docker-compose up -d

setup_redis:
	@cd deployments/compose/dev &&\
	docker-compose up -d redis

teardown_dev: ## Tear down development environment for the emrs project
	@cd deployments/compose/dev &&\
	docker-compose down

redis_console:
	@docker run --rm -it --network bridge-rupa-backend redis redis-cli -h redis

JWT_KEY := hDI0eBv11TbuboZ01qpnOuYRYLh6gQUOQhC9Mfagzv9l3gJso7CalTt7wGzJCVwbeDIfOX6fwS79pnisW7udhQ
API_BLOCK_KEY := 9AI8o4ta02gdqWsVhYe0r276z7my6yDwY78rCsrcofT7pCNq4WwnRoW93hn8WFJM0HheZHDYPc4tD+tUXVNEGw
API_HASH_KEY := 73H/I3+27Qp3ZETqMzbYa/EGT826Zxx2821cmHUl7fTX/DmkIWPJatczkxN3p8RHbdOGWT/HDRAf7gqhZcZOow
FIREBASE_CREDENTIALS_FILE := /Users/jessegitaka/go/src/github.com/gidyon/antibug/service-accout.json
CONFIGS_DIR := /Users/jessegitaka/go/src/github.com/gidyon/services/configs
TEMPLATES_DIR := /Users/jessegitaka/go/src/github.com/gidyon/services/templates/account
ACTIVATION_URL := https://google.com

protoc_account:
	@protoc -I=$(API_IN_PATH) -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) account.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) account.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) account.proto

protoc_messaging:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) messaging.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) messaging.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) messaging.proto
	
protoc_emailing:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) emailing.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) emailing.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) emailing.proto

protoc_push:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) push.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) push.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) push.proto

protoc_sms:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) sms.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) sms.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) sms.proto

protoc_call:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) call.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) call.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) call.proto

protoc_channel:
	@protoc -I=$(API_IN_PATH) -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) channel.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) channel.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) channel.proto

protoc_subscriber:
	@protoc -I=$(API_IN_PATH) -I=third_party --go_out=plugins=grpc:$(API_OUT_PATH) subscriber.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --grpc-gateway_out=logtostderr=true:$(API_OUT_PATH) subscriber.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --swagger_out=logtostderr=true:$(SWAGGER_DOC_OUT_PATH) subscriber.proto

protoc_all: protoc_messaging protoc_emailing protoc_push protoc_sms protoc_call protoc_channel protoc_subscribers

run_account_service:
	cd cmd/services/account && go build -o service && APP_NAME=account_service JWT_SIGNING_KEY=$(JWT_KEY) API_HASH_KEY=$(API_HASH_KEY) API_BLOCK_KEY=$(API_BLOCK_KEY) FIREBASE_CREDENTIALS_FILE=$(FIREBASE_CREDENTIALS_FILE) TEMPLATES_DIR=$(TEMPLATES_DIR) ACTIVATION_URL=$(ACTIVATION_URL) ./service -config-file=$(CONFIGS_DIR)/account.dev.yml

run_channel_service:
	cd cmd/services/channel && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) ./service -config-file=$(CONFIGS_DIR)/channel.dev.yml

run_messaging_service:
	cd cmd/services/messaging && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) SENDER_EMAIL_ADDRESS=gideon.ngk@gmail.com ./service -config-file=$(CONFIGS_DIR)/messaging.dev.yml

run_operations_service:
	cd ./cmd/services/operation && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) ./service -config-file=$(CONFIGS_DIR)/operations.dev.yml

run_subscriber_service:
	cd ./cmd/services/subscriber && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) ./service -config-file=$(CONFIGS_DIR)/subscriber.dev.yml

run_call_service:
	cd cmd/services/messaging/cmd/call && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) ./service -config-file=$(CONFIGS_DIR)/messaging/call.dev.yml

run_email_service:
	cd cmd/services/messaging/cmd/email && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) SMTP_PORT=587 SMTP_HOST=smtp.gmail.com SMTP_USERNAME=emrs.net.ke@gmail.com SMTP_PASSWORD=Haktivah11 ./service -config-file=$(CONFIGS_DIR)/messaging/email.dev.yml

run_push_service:
	cd cmd/services/messaging/cmd/push && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) FCM_SERVER_KEY=AAAAh14hysY:APA91bGrTCdh0Q25aGmGBAjcQxluSrFtvAorlb818qhl6VYmWd2ZRNBJsDMVLB_H2CZ6pTohNXL0WfJtSTbleV8TS8zuXlpHgbCwBwgDqGsXQVGIq5Lrz1hm_AwVIHtFjV77Qc0CFP6o ./service -config-file=$(CONFIGS_DIR)/messaging/push.dev.yml

run_sms_service:
	cd cmd/services/messaging/cmd/sms && go build -v -o service && JWT_SIGNING_KEY=$(JWT_KEY) SMS_API_KEY=abc SMS_AUTH_TOKEN=abc SMS_API_USERNAME=abc SMS_API_PASSWORD=abc SMS_API_URL=abc ./service -config-file=$(CONFIGS_DIR)/messaging/sms.dev.yml

run_apps: run_channel_service run_call_service run_email_service run_push_service run_push_service run_subscriber_service run_account_service run_messaging_service


build_account:
	cd cmd/services/account && make build

build_channel:
	cd cmd/services/channel && make build

build_messaging:
	cd cmd/services/messaging && make build

build_operation:
	cd cmd/services/operation && make build

build_subscriber:
	cd cmd/services/subscriber && make build

build_messaging_email:
	cd cmd/services/messaging/cmd/email && make build

build_messaging_sms:
	cd cmd/services/messaging/cmd/sms && make build

build_messaging_pusher:
	cd cmd/services/messaging/cmd/push && make build

build_messaging_call:
	cd cmd/services/messaging/cmd/call && make build

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
