DOCKER_IMAGE ?= go-trader/go-trader
DOCKER_TAG ?= latest
DOCKER_REF := $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: build
build: 
	go build .

.PHONY: run
run: 
	go build . && DB_MONGO_HOST=localhost:27018 ./go-trader test

.PHONY: test
test: 
	go build . && go test -v ./...

.PHONY: docker
docker:
	docker build -t $(DOCKER_REF) .

.PHONY: compose-up
compose-up:
	@docker-compose up -d

.PHONY: compose-down
compose-down:
	@docker-compose down -v

.PHONY:restart
restart:
	"$(MAKE)"  compose-down
	"$(MAKE)"  compose-up