# Define variables
IMAGE_NAME := golinks
TAG := $(shell \
       LASTTAG=$$(git describe --tags --abbrev=0); \
       COMMITS_SINCE=$$(git rev-list $$LASTTAG..HEAD --count); \
       if [ "$$COMMITS_SINCE" = "0" ]; then \
           echo $$LASTTAG; \
       else \
           echo "$$LASTTAG-dev.$$COMMITS_SINCE-$$(git rev-parse --short HEAD)"; \
       fi)
ARCHS := amd64 arm64
REGISTRY := registry.werewolves.fyi

# Define targets
all: build push-images manifest push

build:
	@echo "Building images for platforms: $(ARCHS)"
	@$(foreach arch, $(ARCHS), \
		docker build --platform linux/$(arch) -t $(REGISTRY)/$(IMAGE_NAME):$(TAG)-$(arch) .;)

push-images:
	@echo "Pushing individual architecture images"
	@$(foreach arch, $(ARCHS), \
		docker push $(REGISTRY)/$(IMAGE_NAME):$(TAG)-$(arch);)

manifest:
	@echo "Creating manifest for images $(REGISTRY)/$(IMAGE_NAME):$(TAG)"
	@docker manifest create $(REGISTRY)/$(IMAGE_NAME):$(TAG) \
		$(foreach arch,$(ARCHS),$(REGISTRY)/$(IMAGE_NAME):$(TAG)-$(arch))

push:
	@echo "Pushing Docker image $(REGISTRY)/$(IMAGE_NAME):$(TAG)"
	@docker manifest push $(REGISTRY)/$(IMAGE_NAME):$(TAG)

clean:
	@echo "Removing built binaries..."
	@rm -rf bin/
	@echo "Removing local Docker image $(IMAGE_NAME):$(TAG)"
	$(foreach arch, $(ARCHS), \
		docker rmi $(IMAGE_NAME):$(TAG)-$(arch);)
	docker manifest rm $(IMAGE_NAME):$(TAG)

run: bin
	@echo "Running golinks container..."
	@mkdir -p /tmp/golinks-config
	@[ -f /tmp/golinks-config/links ] || touch /tmp/golinks-config/links
	@./bin/golinks -level DEBUG -config /tmp/golinks-config/links

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
