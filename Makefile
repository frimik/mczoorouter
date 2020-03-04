NAME = mczoorouter
REGISTRY ?= docker/$(NAME)

# Get the version from .auroraversion file
#VERSION = $(file < .auroraversion)
VERSION = $(shell git describe --tags --dirty --always)

ifeq ($(VERSION),)
	$(error VERSION is not set)
endif


SAFENAME=$(subst /,-,$(NAME))

.PHONY: all
all: build

.PHONY: build
build:
	DOCKER_BUILDKIT=1 \
		docker build -t $(NAME) -f Dockerfile .

release: build
	git checkout releases
	docker create --name $(NAME)-build-results $(NAME):$(VERSION)
	docker cp $(NAME)-build-results:/usr/bin/aurora dist/aurora.pex
	docker rm -f $(NAME)-build-results
	./release.sh


tag: build
	DOCKER_BUILDKIT=1 \
		docker tag $(NAME) $(REGISTRY)
	DOCKER_BUILDKIT=1 \
		docker tag $(NAME) $(REGISTRY):$(VERSION)

push: tag
	DOCKER_BUILDKIT=1 \
		docker push $(REGISTRY)
	DOCKER_BUILDKIT=1 \
		docker push $(REGISTRY):$(VERSION)

tag_latest: build
	DOCKER_BUILDKIT=1 \
		docker tag $(NAME) $(REGISTRY):latest

push_latest: tag_latest
	DOCKER_BUILDKIT=1 \
		docker push $(REGISTRY):latest
