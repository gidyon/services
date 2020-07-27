SERVICE := emailing

compile:
	@GOOS=linux go build -i -v -o service .

docker_build:
ifdef tag
	@docker build -t gidyon/messaging-$(SERVICE):$(tag) .
else
	@docker build -t gidyon/messaging-$(SERVICE):latest .
endif

docker_tag:
ifdef tag
	@docker tag gidyon/messaging-$(SERVICE):$(tag) gidyon/messaging-$(SERVICE):$(tag)
else
	@docker tag gidyon/messaging-$(SERVICE):latest gidyon/messaging-$(SERVICE):latest
endif

docker_push:
ifdef tag
	@docker push gidyon/messaging-$(SERVICE):$(tag)
else
	@docker push gidyon/messaging-$(SERVICE):latest
endif

build_image: docker_build docker_tag docker_push

build: compile docker_build docker_tag docker_push