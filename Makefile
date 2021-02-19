VERSION ?= $(shell cat VERSION)

GITCOMMIT := $(shell git rev-parse HEAD 2>/dev/null)
GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILDTIME := $(shell TZ=GMT date "+%Y-%m-%d_%H:%M_GMT")

SRCS = $(shell find . -name '*.go' | grep -v '^./vendor/')
PKGS := $(foreach pkg, $(sort $(dir $(SRCS))), $(pkg))

TESTARGS ?=

default:
	GO15VENDOREXPERIMENT=1 go build -v

install:
	cp ahoy /usr/local/bin/ahoy
	chmod +x /usr/local/bin/ahoy

cross: dist_dir
	GOOS=linux GOARCH=amd64 GO15VENDOREXPERIMENT=1 \
	  LDFLAGS="-X main.Version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)" \
		go build -v -o ./dist/linux_amd64/ahoy
	
	GOOS=linux GOARCH=arm64 GO15VENDOREXPERIMENT=1 \
		LDFLAGS="-X main.Version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)" \
		go build -v -o ./dist/linux_arm64/ahoy
	
	GOOS=darwin GOARCH=amd64 GO15VENDOREXPERIMENT=1 \
		LDFLAGS="-X main.Version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)" \
		go build -v -o ./dist/darwin_amd64/ahoy
	
	GOOS=darwin GOARCH=arm64 GO15VENDOREXPERIMENT=1 \
		LDFLAGS="-X main.Version=$(VERSION) -X main.GitCommit=$(GITCOMMIT) -X main.GitBranch=$(GITBRANCH) -X main.BuildTime=$(BUILDTIME)" \
		go build -v -o ./dist/darwin_arm64/ahoy

cross_tars: cross
	COPYFILE_DISABLE=1 tar -zcvf ./dist/ahoy_linux_amd64.tar.gz -C dist/linux_amd64 ahoy
	COPYFILE_DISABLE=1 tar -zcvf ./dist/ahoy_linux_arm64.tar.gz -C dist/linux_arm64 ahoy
	COPYFILE_DISABLE=1 tar -zcvf ./dist/ahoy_darwin_amd64.tar.gz -C dist/darwin_amd64 ahoy
	COPYFILE_DISABLE=1 tar -zcvf ./dist/ahoy_darwin_arm64.tar.gz -C dist/darwin_arm64 ahoy

dist_dir:
	mkdir -p ./dist/linux_amd64
	mkdir -p ./dist/linux_arm64
	mkdir -p ./dist/darwin_amd64
	mkdir -p ./dist/darwin_arm64

clean:
	rm -Rf dist

testdeps:
	@ go get github.com/GeertJohan/fgt

fmtcheck:
	$(foreach file,$(SRCS),gofmt $(file) | diff -u $(file) - || exit;)

lint:
	@ go get golang.org/x/lint/golint
	$(foreach file,$(SRCS),fgt golint $(file) || exit;)

vet:
	@ go get golang.org/x/tools/cmd/vet
	$(foreach pkg,$(PKGS),fgt go vet $(pkg) || exit;)

gocyclo:
	@ go get github.com/fzipp/gocyclo
	gocyclo -over 25 ./src

test: testdeps fmtcheck lint vet
	GO15VENDOREXPERIMENT=1 go test ./src/... $(TESTARGS)

version:
	@echo $(VERSION)

.PHONY: clean test fmtcheck lint vet gocyclo version testdeps cross cross_tars dist_dir default install
