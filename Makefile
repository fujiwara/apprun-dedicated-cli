.PHONY: clean test

apprun-dedicated-cli: go.* *.go
	go build -o $@ ./cmd/apprun-dedicated-cli

clean:
	rm -rf apprun-dedicated-cli dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/apprun-dedicated-cli/cmd/apprun-dedicated-cli

dist:
	goreleaser build --snapshot --clean
