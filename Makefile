
GIT_VERSION := $(shell git describe)

.PHONY: package

package:
	@echo "GIT_VERSION = ${GIT_VERSION}"
