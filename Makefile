.DEFAULT_GOAL := install

APP ?= golinks

.PHONY: build
build:
	go build -o bin/

#build-arm:


install:
	go install

images: build-image build-arm-image

build-image:
	buildah build \
		--build-arg BUILDPLATFORM=linux/amd64 \
		--build-arg TARGETPLATFORM=linux/amd64 \
		--build-arg TARGETOS=linux \
		--build-arg TARGETARCH=amd64 \
		--tag $(APP):latest

build-arm-image:
	buildah build \
		--build-arg BUILDPLATFORM=linux/arm \
		--build-arg TARGETPLATFORM=linux/arm \
		--build-arg TARGETOS=linux \
		--build-arg TARGETARCH=arm \
		--tag $(APP):arm

clean:
	rm -rf bin/
