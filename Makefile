setup:
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v
	gometalinter -i -u

build:
	go build

lint:
	gometalinter --vendor --deadline=60s ./...

ci: lint

.DEFAULT_GOAL := build