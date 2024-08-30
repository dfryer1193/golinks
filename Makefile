# Define variables
IMAGE_NAME := dfryer1193/golinks
TAG := latest
PLATFORMS := linux/amd64,linux/arm64
REGISTRY := registry.example.com

# Define targets
all: build push

build:
	@echo "Building Docker images for platforms: $(PLATFORMS)"
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):$(TAG) .

push:
	@echo "Pushing Docker image $(REGISTRY)/$(IMAGE_NAME):$(TAG)"
	docker buildx build --platform $(PLATFORMS) -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) --push .

# Clean up local images
clean:
	@echo "Removing local Docker image $(IMAGE_NAME):$(TAG)"
	docker rmi $(IMAGE_NAME):$(TAG) || true

# List available images
list:
	@echo "Listing Docker images"
	docker images | grep $(IMAGE_NAME)

# Print help
help:
	@echo "Makefile commands:"
	@echo "  make all       - Build and push Docker images"
	@echo "  make build     - Build Docker images"
	@echo "  make push      - Push Docker images to registry"
	@echo "  make clean     - Remove local Docker images"
	@echo "  make list      - List Docker images"
	@echo "  make help      - Show this help message"
