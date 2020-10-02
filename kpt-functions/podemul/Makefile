# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate license fix vet fmt test build tidy image

GOBIN := $(shell go env GOPATH)/bin
GOPATH := $(shell go env GOPATH)

build:
	go build -v -o $(GOBIN)/config-function .

all: generate license build fix vet fmt test lint tidy

fix:
	go fix ./...

fmt:
	go fmt ./...

generate:
	(which $(GOBIN)/mdtogo || go get sigs.k8s.io/kustomize/cmd/mdtogo)
	GOBIN=$(GOBIN) go generate ./...

license:
	(which $(GOBIN)/addlicense || go get github.com/google/addlicense)
	$(GOBIN)/addlicense  -y 2020 -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

tidy:
	go mod tidy

lint:
	(which $(GOBIN)/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1)
	$(GOBIN)/golangci-lint run ./...

test:
	go test -cover ./...

vet:
	go vet ./...

image:
	docker build -f Dockerfile . -t quay.io/aodinokov/podemul:v0.0.1
	#docker push quay.io/aodinokov/podemul:v0.0.1
