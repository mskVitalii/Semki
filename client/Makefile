DATE_TAG := $(shell date +%Y-%m-%d-%H-%M)
VERSION ?= $(DATE_TAG)
IMAGE_NAME = mskkote/semki-ui
PORT = 5173

build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

run:
	docker run --rm -p $(PORT):5173 $(IMAGE_NAME)

push:
	docker build -t $(IMAGE_NAME):$(VERSION) .
	docker push $(IMAGE_NAME):$(VERSION)
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):latest

dev:
	bun run dev
