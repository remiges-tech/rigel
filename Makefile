# Generates a pre-release build from the current commit. Useful for testing and development.
# Artifacts will have a snapshot identifier, and the dist directory will be cleaned before the build.
snapshot:
	goreleaser release --snapshot --rm-dist

# Prepares a release from a tagged commit without publishing it. 
# This allows for manual inspection or testing of artifacts. 
# The dist directory is cleaned before building.
release:
	goreleaser release --skip-publish --rm-dist

.PHONY: snapshot release

