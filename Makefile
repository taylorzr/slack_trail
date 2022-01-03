.DEFAULT_GOAL := help
.PHONY: build clean deploy

help:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/trail

list:
	serverless deploy list functions

deploy: build
ifndef function
	$(error function is undefined)
else
	serverless deploy function --function $(function)
endif

deploy_all: clean build
	serverless deploy --verbose
