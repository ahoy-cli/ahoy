GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
VERSION := $(shell git describe --tags --abbrev=0 $(GITCOMMIT) 2>/dev/null)

GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILDTIME := $(shell TZ=UTC date "+%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := "-s -w -X main.version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)"

# Set binary name based on OS
ifeq ($(OS),Windows_NT)
    BINARY_NAME := ahoy.exe
else
    BINARY_NAME := ahoy
endif

SRCS = $(shell find . -name '*.go' | grep -E -v '^./vendor/')

OS := linux darwin windows
ARCH := amd64 arm64

TESTARGS ?=

default:
	go build -ldflags $(LDFLAGS) -v -o ./$(BINARY_NAME)

install:
	cp ahoy /usr/local/bin/ahoy
	chmod +x /usr/local/bin/ahoy

build_dir:
	mkdir -p ./builds

cross: build_dir
	$(foreach os,$(OS), \
		$(foreach arch,$(ARCH), \
			$(if $(filter windows,$(os)), \
				GOOS=$(os) GOARCH=$(arch) go build -trimpath -ldflags $(LDFLAGS) -v -o ./builds/ahoy-bin-$(os)-$(arch).exe, \
				GOOS=$(os) GOARCH=$(arch) go build -trimpath -ldflags $(LDFLAGS) -v -o ./builds/ahoy-bin-$(os)-$(arch) \
			); \
		) \
	)

clean:
	rm -vRf ./builds/ahoy-bin-*

fmt:
	$(foreach file,$(SRCS),gofmt -w $(file);)

fmtcheck:
	$(foreach file,$(SRCS),gofmt $(file) | diff -u $(file) - || exit;)

staticcheck:
	@ go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

vet:
	go vet ./...

gocyclo:
	@ go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 25 -avg -ignore "vendor" .

# Apply Go code modernizations (uses latest idioms and patterns)
modernize:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix -test ./...

# Check what Go modernizations would be applied (without making changes)
modernize-check:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...

test: fmtcheck staticcheck vet
	go test ./*.go $(TESTARGS)

version:
	@echo $(VERSION)

.PHONY: clean test fmt fmtcheck staticcheck vet gocyclo modernize modernize-check version testdeps cross build_dir default install
