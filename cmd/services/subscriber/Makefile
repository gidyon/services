PROJECT_ROOT := /Users/jessegitaka/go/src/github.com/gidyon/services
SERVICE := subscriber

JWT_KEY := hDI0eBv11TbuboZ01qpnOuYRYLh6gQUOQhC9Mfagzv9l3gJso7CalTt7wGzJCVwbeDIfOX6fwS79pnisW7udhQ
API_BLOCK_KEY := 9AI8o4ta02gdqWsVhYe0r276z7my6yDwY78rCsrcofT7pCNq4WwnRoW93hn8WFJM0HheZHDYPc4tD+tUXVNEGw
API_HASH_KEY := 73H/I3+27Qp3ZETqMzbYa/EGT826Zxx2821cmHUl7fTX/DmkIWPJatczkxN3p8RHbdOGWT/HDRAf7gqhZcZOow

image := gidyon/$(SERVICE):latest
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
	API_BLOCK_KEY=$(API_BLOCK_KEY) \
	API_HASH_KEY=$(API_HASH_KEY) \
	./service -config-file=./config.yaml

gotest:
	@cd $(PROJECT_ROOT)/internal/account && ginkgo -cover
	
compile:
	@GOOS=linux go build -i -v -o service .

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
