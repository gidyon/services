SERVICE := migrator

compile:
	@GOOS=linux CGO_ENABLED=0 go build -v -o $(SERVICE) .

docker_build:
	docker build -t services-migrator .

docker_tag:
	docker tag services-migrator:latest gidyon/services-migrator:latest

docker_push:
	docker push gidyon/services-migrator:latest

build: compile docker_build docker_tag docker_push