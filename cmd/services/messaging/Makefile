SERVICE := messaging

compile:
	@GOOS=linux go build -i -v -o service .

docker_build:
ifdef tag
	@docker build -t gidyon/$(SERVICE):$(tag) .
else
	@docker build -t gidyon/$(SERVICE):latest .
endif

docker_tag:
ifdef tag
	@docker tag gidyon/$(SERVICE):$(tag) gidyon/$(SERVICE):$(tag)
else
	@docker tag gidyon/$(SERVICE):latest gidyon/$(SERVICE):latest
endif

docker_push:
ifdef tag
	@docker push gidyon/$(SERVICE):$(tag)
else
	@docker push gidyon/$(SERVICE):latest
endif

build_image: docker_build docker_tag docker_push

build: compile docker_build docker_tag docker_push