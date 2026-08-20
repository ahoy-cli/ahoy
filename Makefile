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

SRCS = $(shell find . -name '*.go' | grep -v '/vendor/')

OS := linux darwin windows
ARCH := amd64 arm64

TESTARGS ?=

default:
	go build -ldflags $(LDFLAGS) -v -o ./$(BINARY_NAME)

install:
	cp $(BINARY_NAME) /usr/local/bin/ahoy
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

# Mirrors the Lint step in .github/workflows/build_and_test.yml. Deliberately
# not part of `test` while pre-existing complexity findings are outstanding;
# CI runs it with continue-on-error for the same reason.
lint:
	@ command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install it: https://golangci-lint.run/docs/welcome/install/"; \
		exit 1; \
	}
	golangci-lint run

test: fmtcheck staticcheck vet
	go test ./*.go $(TESTARGS)

version:
	@echo $(VERSION)

.PHONY: clean test fmtcheck staticcheck vet lint gocyclo version testdeps cross build_dir default install
