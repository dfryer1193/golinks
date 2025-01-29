# Define variables
IMAGE_NAME := decahedra/golinks
TAG := latest
ARCHS := amd64 arm64
REGISTRY := docker.io

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
	@echo "Removing local Docker image $(IMAGE_NAME):$(TAG)"
	$(foreach arch, $(ARCHS), \
		docker rmi $(IMAGE_NAME):$(TAG)-$(arch);)
	docker manifest rm $(IMAGE_NAME):$(TAG)

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
	@echo "  make clean     - Remove local Docker images"
	@echo "  make list      - List Docker images"
	@echo "  make help      - Show this help message"
