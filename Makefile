API_IN_PATH := api/proto
API_OUT_PATH := pkg/api
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

protoc_longrunning:
	@protoc -I=$(API_IN_PATH) -I=third_party --go-grpc_out=$(API_OUT_PATH)/longrunning --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/longrunning longrunning.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/longrunning longrunning.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) longrunning.proto

protoc_account:
	@protoc -I=$(API_IN_PATH) -I=$(API_IN_PATH)/messaging -I=third_party --go-grpc_out=$(API_OUT_PATH)/account --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/account account.proto
	@protoc -I=$(API_IN_PATH) -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/account account.proto
	@protoc -I=$(API_IN_PATH) -I=$(API_IN_PATH)/messaging -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) account.proto

protoc_messaging:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go-grpc_out=$(API_OUT_PATH)/messaging --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/messaging messaging.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/messaging messaging.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) messaging.proto
	
protoc_emailing:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go-grpc_out=$(API_OUT_PATH)/messaging/emailing --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/messaging/emailing emailing.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/messaging/emailing emailing.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) emailing.proto

protoc_pusher:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go-grpc_out=$(API_OUT_PATH)/messaging/pusher --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/messaging/pusher pusher.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/messaging/pusher pusher.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) pusher.proto

protoc_sms:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go-grpc_out=$(API_OUT_PATH)/messaging/sms --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/messaging/sms sms.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/messaging/sms sms.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) sms.proto

protoc_call:
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --go-grpc_out=$(API_OUT_PATH)/messaging/call --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/messaging/call call.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/messaging/call call.proto
	@protoc -I=$(API_IN_PATH)/messaging -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) call.proto

protoc_channel:
	@protoc -I=$(API_IN_PATH) -I=third_party --go-grpc_out=$(API_OUT_PATH)/channel --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/channel channel.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/channel channel.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) channel.proto

protoc_subscriber:
	@protoc -I=$(API_IN_PATH) -I=third_party --go-grpc_out=$(API_OUT_PATH)/subscriber --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/subscriber subscriber.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/subscriber subscriber.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) subscriber.proto

protoc_settings:
	@protoc -I=$(API_IN_PATH) -I=third_party --go-grpc_out=$(API_OUT_PATH)/settings --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/settings settings.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --grpc-gateway_out=logtostderr=true,paths=source_relative:$(API_OUT_PATH)/settings settings.proto
	@protoc -I=$(API_IN_PATH) -I=third_party --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:$(SWAGGER_DOC_OUT_PATH) settings.proto

protoc_error:
	@protoc -I=$(API_IN_PATH) -I=third_party --go-grpc_out=$(API_OUT_PATH)/usererror --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=$(API_OUT_PATH)/usererror error.proto

protoc_all: protoc_account protoc_messaging protoc_emailing protoc_pusher protoc_sms protoc_call protoc_channel protoc_subscriber protoc_settings protoc_longrunning

cp_doc:
	@cp -r $(SWAGGER_DOC_OUT_PATH)/ cmd/apidoc/dist/swagger/
	
gen_api_doc: protoc_all cp_doc


run_account_service:
	cd cmd/services/account && make run

run_channel_service:
	cd cmd/services/channel && make run

run_messaging_service:
	cd cmd/services/messaging && make run

run_longrunning_service:
	cd ./cmd/services/longrunning && make run

run_subscriber_service:
	cd ./cmd/services/subscriber && make run

run_call_service:
	cd cmd/services/messaging/cmd/call && make run

run_email_service:
	cd cmd/services/emailing && make run

run_push_service:
	cd cmd/services/pusher && make run

run_sms_service:
	cd cmd/services/sms && make run

run_all: run_account_service run_call_service run_channel_service run_email_service run_longrunning_service run_messaging_service run_push_service run_sms_service run_subscriber_service


build_account:
	cd cmd/services/account && sudo make build_service

build_channel:
	cd cmd/services/channel && sudo make build_service

build_messaging:
	cd cmd/services/messaging && sudo make build_service

build_longrunning:
	cd cmd/services/longrunning && sudo make build_service

build_subscriber:
	cd cmd/services/subscriber && sudo make build_service

build_messaging_email:
	cd cmd/services/emailing && sudo make build_service

build_messaging_sms:
	cd cmd/services/sms && sudo make build_service

build_messaging_pusher:
	cd cmd/services/pusher && sudo make build_service

build_messaging_call:
	cd cmd/services/call && sudo make build_service

build_all: build_account build_messaging_call build_messaging_email build_longrunning build_messaging build_messaging_pusher build_messaging_sms build_subscriber build_channel

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
