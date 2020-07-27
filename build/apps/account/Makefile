PROJECT_NAME := emrs
PKG := github.com/gidyon/$(PROJECT_NAME)
SERVICE_CMD_FOLDER := ${PKG}/cmd

compile_binary:
	go build -i -v -o account $(SERVICE_CMD_FOLDER)/apps/account

docker_build: ## Create a docker image for the service
ifdef tag
	@docker build -t gidyon/$(PROJECT_NAME)-account:$(tag) .
else
	@docker build -t gidyon/$(PROJECT_NAME)-account:latest .
endif

docker_tag:
ifdef tag
	@docker tag gidyon/$(PROJECT_NAME)-account:$(tag) gidyon/$(PROJECT_NAME)-account:$(tag)
else
	@docker tag gidyon/$(PROJECT_NAME)-account:latest gidyon/$(PROJECT_NAME)-account:latest
endif

docker_push:
ifdef tag
	@docker push gidyon/$(PROJECT_NAME)-account:$(tag)
else
	@docker push gidyon/$(PROJECT_NAME)-account:latest
endif

build_image: docker_build docker_tag docker_push

build: compile_binary docker_build docker_tag docker_push