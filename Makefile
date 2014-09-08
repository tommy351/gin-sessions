deps:
	go get github.com/tools/godep
	godep restore

install: deps

test: export GO_ENV=test
test:
	godep go test -v