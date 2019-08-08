.PHONY: build clean deploy gomodgen

build: gomodgen
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/users users/*.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/emojis emojis/*.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose --aws-profile personal

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh
