PROJECT_ROOT := /Users/jessegitaka/go/src/github.com/gidyon/services
SERVICE := messaging-sms

JWT_KEY := hDI0eBv11TbuboZ01qpnOuYRYLh6gQUOQhC9Mfagzv9l3gJso7CalTt7wGzJCVwbeDIfOX6fwS79pnisW7udhQ

image := gidyon/$(SERVICE):v2
context := .

ifdef IMAGE
	image=$(IMAGE)
endif

ifdef BUILD_CONTEXT
	context=$(BUILD_CONTEXT)
endif
run:
	go build -v -o service && \
	JWT_SIGNING_KEY=$(JWT_KEY) \
	SMS_API_PASSWORD=abc \
	SMS_API_URL=abc \
	SENDER_ID=22141 \
	SMS_API_KEY=eGh/LkniEmyaruyee6aQYCNbvdYwx6sUNr6CsSkORWM= \
	SMS_API_USERNAME=0b3c6666-5d42-4fdf-90d0-be8bec00e665 \
	SMS_AUTH_TOKEN=7SbwNHZNhGUUg1FfIxKNr7oKzbUgNGSQ \
	SMS_ACCESS_TOKEN=2MJcjwjwOFMTAFep4iYV4jzMWybkxlXg \
	./service -config-file=./config.yaml

gotest:
	@cd $(PROJECT_ROOT)/internal/messaging/call && ginkgo -cover
	
compile:
	@GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -v -o service .

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
