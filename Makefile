.PHONY: build clean deploy gomodgen

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/trail

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
ifndef function
	$(error function is undefined)
else
	serverless deploy function --function $(function)
endif

deploy_all: clean build
	serverless deploy --verbose
