
IMG_REPOSITORY ?= af.hikvision.com.cn/docker-drpd
# Image URL to use all building/pushing image targets
IMG ?= ${IMG_REPOSITORY}/kubernetes/device-manager:v0.1

PROJECTPATH ?= hikvision.com/cloud/device-manager
GITVERSION ?= $(shell git describe --abbrev=0 --tags)
GITCOMMIT ?= $(shell git rev-parse HEAD)
DATE ?= $(shell date +'%Y-%m-%dT%H:%M:%SZ')

fmt:
	go fmt ./pkg/... ./cmd/...

docker-build:
	docker run --rm -it -v $(GOPATH)/src/hikvision.com/cloud/device-manager:/go/src/hikvision.com/cloud/device-manager \
		--workdir /go/src/hikvision.com/cloud/device-manager \
		af.hikvision.com.cn/docker-drpd/library/golang:1.13.3-tools \
		go build -mod=vendor -ldflags "-X $(PROJECTPATH)/pkg/version.GitVersion=$(GITVERSION) -X $(PROJECTPATH)/pkg/version.GitCommit=$(GITCOMMIT) -X $(PROJECTPATH)/pkg/version.BuildDate=$(DATE)" \
		-o bin/device-manager ./cmd

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	-ldflags "-X $(PROJECTPATH)/pkg/version.GitVersion=$(GITVERSION) -X $(PROJECTPATH)/pkg/version.GitCommit=$(GITCOMMIT) -X $(PROJECTPATH)/pkg/version.BuildDate=$(DATE)" \
	-o bin/device-manager ./cmd

build-cli:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    	-ldflags "-X $(PROJECTPATH)/pkg/version.GitVersion=$(GITVERSION) -X $(PROJECTPATH)/pkg/version.GitCommit=$(GITCOMMIT) -X $(PROJECTPATH)/pkg/version.BuildDate=$(DATE)" \
    	-o bin/device-manager-cli ./cli

docker-image:
	sudo docker build -t ${IMG} .

docker-lint:
	docker run --rm -it -v $GOPATH/src/hikvision.com/cloud/device-manager:/go/src/hikvision.com/cloud/device-manager \
		--workdir /go/src/hikvision.com/cloud/device-manager \
		-e GOFLAGS=-mod=vendor \
		af.hikvision.com.cn/docker-drpd/library/golang:1.13.3-tools \
		golint -set_exit_status $(go list ./... | grep -v /vendor/)

golint:
	golint $(shell go list ./... | grep -v /vendor/)

generate:
	$(GOPATH)/src/k8s.io/code-generator/generate-groups.sh deepcopy,defaulter,client \
	hikvision.com/cloud/device-manager/pkg/crd/client \
	hikvision.com/cloud/device-manager/pkg/crd/apis device.k8s.io:v1alpha1