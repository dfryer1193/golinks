# Define variables
IMAGE_NAME := decahedra/golinks
TAG := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))
ARCHS := amd64 arm64
REGISTRY := docker.io
CURRENT_ARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')

# Define targets
all: build manifest push

build:
	@echo "Building images for platforms: $(ARCHS)"
	@$(foreach arch, $(ARCHS), \
		docker build --arch $(arch) -t $(IMAGE_NAME):$(TAG)-$(arch) .;)

manifest:
	@echo "Creating manifest for images"
	@docker manifest create $(IMAGE_NAME):$(TAG)
	@$(foreach arch, $(ARCHS), \
		docker manifest add $(IMAGE_NAME):$(TAG) containers-storage:localhost/$(IMAGE_NAME):$(TAG)-$(arch);)

push:
	@echo "Pushing Docker image $(REGISTRY)/$(IMAGE_NAME):$(TAG)"
	@docker manifest push --all $(IMAGE_NAME):$(TAG) docker://$(REGISTRY)/$(IMAGE_NAME):$(TAG)

clean:
	@echo "Removing built binaries..."
	@rm -rf bin/
	@echo "Removing local Docker image $(IMAGE_NAME):$(TAG)"
	$(foreach arch, $(ARCHS), \
		docker rmi $(IMAGE_NAME):$(TAG)-$(arch);)
	docker manifest rm $(IMAGE_NAME):$(TAG)

run: build
	@echo "Running golinks container..."
	@mkdir -p /tmp/golinks-config
	@[ -f /tmp/golinks-config/links ] || touch /tmp/golinks-config/links
	@docker run --rm \
		--name golinks \
		-p 8080:8080 \
		-v /tmp/golinks-config:/config:rw \
		localhost/$(IMAGE_NAME):$(TAG)-$(CURRENT_ARCH)
	@echo "Golinks is running on http://localhost:8080"

bin:
	@echo "Building golinks..."
	@go build -o bin/ cmd/golinks/golinks.go

install:
	@echo "Installing golinks..."
	@go install cmd/golinks/golinks.go

list:
	@echo "Listing Docker images"
	docker images | grep $(IMAGE_NAME)

# Print help
help:
	@echo "Makefile commands:"
	@echo "  make all       - Build and push Docker images"
	@echo "  make build     - Build Docker images"
	@echo "  make manifest  - Create a manifest for built images"
	@echo "  make push      - Push Docker images to registry"
	@echo "  make bin       - Build binary"
	@echo "  make install   - Install binary"
	@echo "  make clean     - Remove local binaries and Docker images"
	@echo "  make list      - List Docker images"
	@echo "  make run       - Run golinks container (Ctrl+C to stop)"
	@echo "  make help      - Show this help message"
