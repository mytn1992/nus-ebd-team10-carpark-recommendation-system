.PHONY: build clean deploy

run:
	export GO111MODULE=on
	go run cmd/post/main.go

run-random:
	export GO111MODULE=on
	go run cmd/random/main.go

run-random-elastic:
	export GO111MODULE=on
	go run cmd/random_elastic/main.go
	
build:
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/post cmd/post/main.go

clean:
	rm -rf ./bin ./vendor go.sum

deploy: build
	sls deploy --verbose
