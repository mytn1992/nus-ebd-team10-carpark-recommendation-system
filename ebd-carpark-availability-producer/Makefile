.PHONY: build clean deploy gomodgen

run:
	export GO111MODULE=on && go run cmd/carpark_availability/main.go 
run2:
	export GO111MODULE=on && go run cmd/carpark_information/main.go 
run-weather:
	export GO111MODULE=on && go run cmd/weather/main.go 
run-query:
	export GO111MODULE=on && go run cmd/query/main.go 
run-processed:
	export GO111MODULE=on && go run cmd/push-processed/main.go 
run-extract:
	export GO111MODULE=on && go run cmd/extract/main.go 

test:
	export GO111MODULE=on && go test -v ./...

build-kafka: 
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -tags musl -ldflags="-s -w" -o bin/carpark_availability cmd/carpark_availability/main.go

build: 
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/carpark_information cmd/carpark_information/main.go
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/weather cmd/weather/main.go
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/extract cmd/extract/main.go
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/process cmd/push-processed-s3/main.go

build2: 
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/carpark_information cmd/carpark_information/main.go
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/weather cmd/weather/main.go
	
deploy:
	sls deploy --verbose --force
