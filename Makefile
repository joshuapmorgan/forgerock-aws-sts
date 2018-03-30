ORG_PATH = github.com/joshuapmorgan
BINARY_NAME = forgerock-aws-sts
REPO_PATH = $(ORG_PATH)/$(BINARY_NAME)

PLATFORMS = darwin/amd64 linux/amd64 windows/amd64
TEMP = $(subst /, , $@)
OS = $(word 1, $(TEMP))
ARCH = $(word 2, $(TEMP))

VERSION_VAR = $(REPO_PATH)/version.Version
GIT_VAR = $(REPO_PATH)/version.GitCommit
BUILD_DATE_VAR = $(REPO_PATH)/version.BuildDate

REPO_VERSION = $$(git describe --abbrev=0 --tags)
BUILD_DATE = $$(date +%Y%m%d-%H%M)
GIT_HASH = $$(git rev-parse --short HEAD)

GOBUILD_VERSION_ARGS := -ldflags "-s -X $(VERSION_VAR)=$(REPO_VERSION) -X $(GIT_VAR)=$(GIT_HASH) -X $(BUILD_DATE_VAR)=$(BUILD_DATE)"

GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

.PHONY: $(PLATFORMS)
$(PLATFORMS): *.go fmt
	GOOS=$(OS) GOARCH=$(ARCH) go build -o build/bin/$(ARCH)/$(OS)/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS) $(REPO_PATH)

.PHONY: build
build: $(PLATFORMS)

.PHONY: fmt
fmt:
	gofmt -w $(GOFMT_FILES)

.PHONY: version
version:
	@echo $(REPO_VERSION)
