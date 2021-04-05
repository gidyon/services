PROJECT_ROOT := /Users/jessegitaka/go/src/github.com/gidyon/services
SERVICE := messaging-pusher

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
	FCM_SERVER_KEY=AAAAh14hysY:APA91bGrTCdh0Q25aGmGBAjcQxluSrFtvAorlb818qhl6VYmWd2ZRNBJsDMVLB_H2CZ6pTohNXL0WfJtSTbleV8TS8zuXlpHgbCwBwgDqGsXQVGIq5Lrz1hm_AwVIHtFjV77Qc0CFP6o \
	JWT_SIGNING_KEY=$(JWT_KEY) \
	./service -config-file=./config.yaml

gotest:
	@cd $(PROJECT_ROOT)/internal/messaging/pusher && ginkgo -cover
	
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
