PROJECT_NAME := emrs
PKG := github.com/gidyon/$(PROJECT_NAME)
SERVICE_CMD_FOLDER := ${PKG}/cmd

compile_binary:
	go build -i -v -o permission $(SERVICE_CMD_FOLDER)/chaincodes/permission

docker_build: ## Create a docker image for the service
ifdef tag
	@docker build -t gidyon/$(PROJECT_NAME)-permission-chaincode:$(tag) .
else
	@docker build -t gidyon/$(PROJECT_NAME)-permission-chaincode:latest .
endif

docker_tag:
ifdef tag
	@docker tag gidyon/$(PROJECT_NAME)-permission-chaincode:$(tag) gidyon/$(PROJECT_NAME)-permission-chaincode:$(tag)
else
	@docker tag gidyon/$(PROJECT_NAME)-permission-chaincode:latest gidyon/$(PROJECT_NAME)-permission-chaincode:latest
endif

docker_push:
ifdef tag
	@docker push gidyon/$(PROJECT_NAME)-permission-chaincode:$(tag)
else
	@docker push gidyon/$(PROJECT_NAME)-permission-chaincode:latest
endif

build: compile_binary docker_build docker_tag docker_push