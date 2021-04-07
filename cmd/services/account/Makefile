PROJECT_ROOT := /Users/jessegitaka/go/src/github.com/gidyon/services
SERVICE := account

JWT_KEY := hDI0eBv11TbuboZ01qpnOuYRYLh6gQUOQhC9Mfagzv9l3gJso7CalTt7wGzJCVwbeDIfOX6fwS79pnisW7udhQ
API_BLOCK_KEY := 9AI8o4ta02gdqWsVhYe0r276z7my6yDwY78rCsrcofT7pCNq4WwnRoW93hn8WFJM0HheZHDYPc4tD+tUXVNEGw
API_HASH_KEY := 73H/I3+27Qp3ZETqMzbYa/EGT826Zxx2821cmHUl7fTX/DmkIWPJatczkxN3p8RHbdOGWT/HDRAf7gqhZcZOow
FIREBASE_CREDENTIALS_FILE := ~/go/src/bitbucket.org/gideonkamau/ussd-gateway/onfonportal-firebase-adminsdk-tt747-416aa4a470.json
CONFIGS_DIR := /Users/jessegitaka/go/src/github.com/gidyon/services/configs
TEMPLATES_DIR := /home/gideon/go/src/github.com/gidyon/services/templates/account
ACTIVATION_URL := https://google.com
EMAIL_DISPLAY_NAME := "Account Service"

image := gidyon/$(SERVICE)
context := .

ifdef IMAGE
	image=$(IMAGE)
else
	imagex := $(image)
	ifdef tag
		image=$(imagex):$(tag)
	else
		image=$(imagex):v2
	endif
endif

ifdef BUILD_CONTEXT
	context=$(BUILD_CONTEXT)
endif

run:
	GOOS=linux go build -v -o service && \
	JWT_SIGNING_KEY=$(JWT_KEY) \
	APP_NAME=account_service \
	FIREBASE_CREDENTIALS_FILE=$(FIREBASE_CREDENTIALS_FILE) \
	TEMPLATES_DIR=$(TEMPLATES_DIR) \
	ACTIVATION_URL=$(ACTIVATION_URL) \
	EMAIL_DISPLAY_NAME=$(EMAIL_DISPLAY_NAME) \
	./service -config-file=./config.yaml

gotest:
	@cd $(PROJECT_ROOT)/internal/account && ginkgo -cover
	
compile:
	@GOOS=linux CGO_ENABLED=0 go build -tags netgo -installsuffix netgo -v -o service .

docker_build:
	@docker build -t $(image) .

docker_tag:
	@docker tag $(image) $(image)

docker_push:
	@docker push $(image)

build_service: compile docker_build docker_tag docker_push

deploy:
	@kubectl delete -f deploy.yaml && kubectl apply -f deploy.yaml

build_and_deploy: gotest compile docker_build docker_tag docker_push deploy
