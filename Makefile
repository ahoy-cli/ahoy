GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
VERSION := $(shell git describe --tags --abbrev=0 $(GITCOMMIT) 2>/dev/null)

GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILDTIME := $(shell TZ=UTC date "+%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := "-s -w -X main.version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)"

SRCS = $(shell find . -name '*.go' | grep -v '^./vendor/')
PKGS := $(foreach pkg, $(sort $(dir $(SRCS))), $(pkg))

OS := linux darwin windows
ARCH := amd64 arm64

TESTARGS ?=

default:
  cd v2 && go build -ldflags $(LDFLAGS) -v -o ./ahoy

install:
	cp ./v2/ahoy /usr/local/bin/ahoy
	chmod +x /usr/local/bin/ahoy

build_dir:
	mkdir -p ./v2/builds

cross: build_dir
	cd v2
	$(foreach os,$(OS), \
		$(foreach arch,$(ARCH), \
			GOOS=$(os) GOARCH=$(arch) go build -trimpath -ldflags $(LDFLAGS) -v -o ./builds/ahoy-bin-$(os)-$(arch); \
		) \
	)

	$(foreach arch,$(ARCH),mv ./builds/ahoy-bin-windows-$(arch) ./builds/ahoy-bin-windows-$(arch).exe;)

clean:
	rm -vRf ./v2/builds/ahoy-bin-*

fmtcheck:
	$(foreach file,$(SRCS),gofmt $(file) | diff -u $(file) - || exit;)

staticcheck:
	@ go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

vet:
	$(foreach pkg,$(PKGS),go vet $(pkg) || exit;)

gocyclo:
	@ go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 25 -avg -ignore "vendor" ./v2

test: fmtcheck staticcheck vet
	 go test ./v2/*.go $(TESTARGS)

version:
	@echo $(VERSION)

.PHONY: clean test fmtcheck staticcheck vet gocyclo version testdeps cross build_dir default install
