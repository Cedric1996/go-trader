DOCKER_IMAGE ?= go-trader/go-trader
DOCKER_TAG ?= latest
DOCKER_REF := $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: run
run: 
	go build . && ./go-trader test

.PHONY: docker
docker:
	docker build -t $(DOCKER_REF) .

.PHONY: compose-up
compose-up:
	@docker-compose up -d