SERVICE := ussdapidoc

compile:
	@GOOS=linux go build -i -v -o service .

docker_build:
ifdef tag
	@docker build -t onfon/$(SERVICE):$(tag) .
else
	@docker build -t onfon/$(SERVICE):latest .
endif

docker_tag:
ifdef tag
	@docker tag onfon/$(SERVICE):$(tag) onfon/$(SERVICE):$(tag)
else
	@docker tag onfon/$(SERVICE):latest onfon/$(SERVICE):latest
endif

docker_push:
ifdef tag
	@docker push onfon/$(SERVICE):$(tag)
else
	@docker push onfon/$(SERVICE):latest
endif

build_image: docker_build docker_tag docker_push

build: compile docker_build docker_tag docker_push