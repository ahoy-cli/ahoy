GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
VERSION := $(shell git describe --tag $(GITCOMMIT) 2>/dev/null)

GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILDTIME := $(shell TZ=GMT date "+%Y-%m-%d_%H:%M_GMT")
LDFLAGS := "-s -w -X main.version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)"

SRCS = $(shell find . -name '*.go' | grep -v '^./vendor/')
PKGS := $(foreach pkg, $(sort $(dir $(SRCS))), $(pkg))

OS := linux darwin windows
ARCH := amd64 arm64

TESTARGS ?=

default:
	go build -ldflags $(LDFLAGS) -v -o ./ahoy

install:
	cp ahoy /usr/local/bin/ahoy
	chmod +x /usr/local/bin/ahoy

build_dir:
	mkdir -p ./builds

cross: build_dir
	$(foreach os,$(OS), \
		$(foreach arch,$(ARCH), \
			GOOS=$(os) GOARCH=$(arch) go build -trimpath -ldflags $(LDFLAGS) -v -o ./builds/ahoy-bin-$(os)-$(arch); \
		) \
	)

	$(foreach arch,$(ARCH),mv ./builds/ahoy-bin-windows-$(arch) ./builds/ahoy-bin-windows-$(arch).exe;)

clean:
	rm -vRf ./builds/ahoy-bin-*

fmtcheck:
	$(foreach file,$(SRCS),gofmt $(file) | diff -u $(file) - || exit;)

staticcheck:
	@ go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

vet:
	$(foreach pkg,$(PKGS),go vet $(pkg) || exit;)

gocyclo:
	@ go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 25 -avg -ignore "vendor" .

test: fmtcheck staticcheck vet
	 go test *.go $(TESTARGS)

version:
	@echo $(VERSION)

.PHONY: clean test fmtcheck staticcheck vet gocyclo version testdeps cross build_dir default install
