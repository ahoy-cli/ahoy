VERSION ?= $(shell cat VERSION)

GITCOMMIT := $(shell git rev-parse HEAD 2>/dev/null)
GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILDTIME := $(shell TZ=GMT date "+%Y-%m-%d_%H:%M_GMT")
LDFLAGS := "-X main.Version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)"

SRCS = $(shell find . -name '*.go' | grep -v '^./vendor/')
PKGS := $(foreach pkg, $(sort $(dir $(SRCS))), $(pkg))

TESTARGS ?=

default:
	go build -v

install:
	cp ahoy /usr/local/bin/ahoy
	chmod +x /usr/local/bin/ahoy

cross: build_dir
	GOOS=linux GOARCH=amd64 \
	  LDFLAGS=$(LDFLAGS) \
		go build -v -o ./builds/linux_amd64/ahoy
	
	GOOS=linux GOARCH=arm64 \
		LDFLAGS=$(LDFLAGS) \
		go build -v -o ./builds/linux_arm64/ahoy
	
	GOOS=darwin GOARCH=amd64  \
		LDFLAGS=$(LDFLAGS) \
		go build -v -o ./builds/darwin_amd64/ahoy
	
	GOOS=darwin GOARCH=arm64  \
		LDFLAGS=$(LDFLAGS) \
		go build -v -o ./builds/darwin_arm64/ahoy

cross_tars: cross
	COPYFILE_DISABLE=1 tar -zcvf ./builds/ahoy_linux_amd64.tar.gz -C builds/linux_amd64 ahoy
	COPYFILE_DISABLE=1 tar -zcvf ./builds/ahoy_linux_arm64.tar.gz -C builds/linux_arm64 ahoy
	COPYFILE_DISABLE=1 tar -zcvf ./builds/ahoy_darwin_amd64.tar.gz -C builds/darwin_amd64 ahoy
	COPYFILE_DISABLE=1 tar -zcvf ./builds/ahoy_darwin_arm64.tar.gz -C builds/darwin_arm64 ahoy

build_dir:
	mkdir -p ./builds/linux_amd64
	mkdir -p ./builds/linux_arm64
	mkdir -p ./builds/darwin_amd64
	mkdir -p ./builds/darwin_arm64

clean:
	cd builds
	rm -Rf linux* darwin* *.tar.gz

fmtcheck:
	$(foreach file,$(SRCS),gofmt $(file) | diff -u $(file) - || exit;)

lint:
	@ go get golang.org/x/lint/golint
	$(foreach file,$(SRCS),golint $(file) || exit;)

vet:
	$(foreach pkg,$(PKGS),go vet $(pkg) || exit;)

gocyclo:
	@ go get github.com/fzipp/gocyclo/cmd/gocyclo
	gocyclo -over 25 -avg -ignore "vendor" .

test: fmtcheck lint vet
	 go test *.go $(TESTARGS)

version:
	@echo $(VERSION)

.PHONY: clean test fmtcheck lint vet gocyclo version testdeps cross cross_tars build_dir default install
