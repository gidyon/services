PROJECT_ROOT := /Users/jessegitaka/go/src/github.com/gidyon/services
SERVICE := projects-api
JWT_KEY := hDI0eBv11TbuboZ01qpnOuYRYLh6gQUOQhC9Mfagzv9l3gJso7CalTt7wGzJCVwbeDIfOX6fwS79pnisW7udhQ

image := gidyon/$(SERVICE)
context := .

ifdef IMAGE
	image=$(IMAGE)
else
	imagex := $(image)
	ifdef tag
		image=$(imagex):$(tag)
	else
		image=$(imagex):latest
	endif
endif

ifdef BUILD_CONTEXT
	context=$(BUILD_CONTEXT)
endif

run:
	GOOS=linux go build -v -o service && \
	JWT_SIGNING_KEY=$(JWT_KEY) \
	./service -config-file=./config.yaml

gotest:
	@cd $(PROJECT_ROOT)/internal/project/v1 && ginkgo -cover
	
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
