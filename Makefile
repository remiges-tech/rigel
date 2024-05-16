.PHONY: build-wsc-dev run-wsc-dev build-rigelctl build-docker snapsho release

build-wsc-dev:
	mkdir -p out
	cd server && go build -tags dev -o ../out/rigel-server .

run-wsc-dev: build-wsc-dev
	cd server && ../out/rigel-server

build-rigelctl:
	mkdir -p out
	cd cmd/rigelctl && go build -o ../../out/rigelctl .


# Default build tag is empty for non-dev builds
BUILD_TAGS ?=

# Build container image for rigel web services server with optional build tags
# for dev build
#     make build-docker BUILD_TAGS=dev
# for non-dev build
#     make build-docker
build-docker:
	docker build --build-arg BUILD_TAGS=$(BUILD_TAGS) -t rigelwsc:latest .
# Generates a pre-release build from the current commit. Useful for testing and development.
# Artifacts will have a snapshot identifier, and the dist directory will be cleaned before the build.
snapshot:
	goreleaser release --snapshot --rm-dist

# Prepares a release from a tagged commit without publishing it. 
# The binaries will be published to the Github release manually.
# The dist directory is cleaned before building.
release:
	goreleaser release --skip-publish --rm-dist

