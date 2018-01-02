IMG ?= sheens
NAMESPACE ?= williamofockham
TAG ?= latest
CONTAINER ?= sheens-build
DOCKER_SOCKET_ARG := $(if $(DOCKER_HOST), -e DOCKER_HOST=$(DOCKER_HOST), -v \
/var/run/docker.sock:/var/run/docker.sock)
LINUX_HEADERS = -v /lib/modules:/lib/modules

.PHONY: build push pull image run rmi tag

build:
	@docker build -t $(IMG):$(TAG) .

run:
	@docker run --name $(CONTAINER) -it --rm --pid='host' \
	-v `pwd`:/$(CONTAINER) $(LINUX_HEADERS) ${DOCKER_SOCKET_ARG} \
	$(IMG) $(DOCKER_RUN_ARG)

tag:
	@docker tag $(IMG) $(NAMESPACE)/$(IMG):$(TAG)

push:
	@docker push $(NAMESPACE)/$(IMG):$(TAG)

pull:
	@docker pull $(NAMESPACE)/$(IMG):$(TAG)

image: build tag push

rmi:
	@docker rmi $(IMG):$(TAG)
	@docker rmi $(NAMESPACE)/$(IMG):$(TAG)
