SHELL := $(shell which bash) # set default shell

# OS / Arch we will build our binaries for
OSARCH := "linux/amd64 linux/386 windows/amd64 windows/386"
ENV = /usr/bin/env

.SHELLFLAGS = -c # Run commands in a -c flag
.SILENT: ; # no need for @
.ONESHELL: ; # recipes execute in same shell
.NOTPARALLEL: ; # wait for this target to finish
.EXPORT_ALL_VARIABLES: ; # send all vars to shell

.PHONY: all # All targets are accessible for user
.DEFAULT: help # Running Make will run the help target

help: ## Show Help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

dep: ## Get build dependencies
	go get -v -u github.com/golang/dep/cmd/dep && \
	go get github.com/mitchellh/gox && \
	go get github.com/mattn/goveralls

build: ## Build the app
	dep ensure && cd cmd/readordie && go build

cross-build: ## Build the app for multiple os/arch
	gox -osarch=$(OSARCH) -output "bin/readordie_{{.OS}}_{{.Arch}}"

test: ## Launch tests
	go test -v ./...