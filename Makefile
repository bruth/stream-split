PROG ?= stream-split
IMAGE ?= bruth/$(PROG)

COMMIT ?= $(shell git rev-parse --short HEAD)
BRANCH ?= $(shell git symbolic-ref -q --short HEAD)
TAG ?= $(shell git describe --tags --exact-match 2>/dev/null)

GO111MODULE ?= on
GOPATH ?= $(shell go env GOPATH)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build:
	go build -o $(CURDIR)/dist/$(GOOS)-$(GOARCH)/$(PROG) $(CURDIR)

image:
	docker build -t $(IMAGE):$(COMMIT) .
ifeq ($(BRANCH),master)
	docker tag $(IMAGE):$(COMMIT) $(IMAGE):latest
else ifneq ($(BRANCH),)
	docker tag $(IMAGE):$(COMMIT) $(IMAGE):$(BRANCH)
endif
ifneq ($(TAG),)
	docker tag $(IMAGE):$(COMMIT) $(IMAGE):$(TAG)
endif

push:
	docker $(PUSHARGS) push $(IMAGE):$(COMMIT)
ifeq ($(BRANCH),master)
	docker $(PUSHARGS) push $(IMAGE):latest
else ifneq ($(BRANCH),)
	docker $(PUSHARGS) push $(IMAGE):$(BRANCH)
endif
ifneq ($(TAG),)
	docker $(PUSHARGS) push $(IMAGE):$(TAG)
endif

.PHONY: build image push
