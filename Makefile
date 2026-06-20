.PHONY: clean test

BUILD_TAGS = no_gcs,no_azurerm

apprun-dedicated-cli: go.* *.go
	go build -tags $(BUILD_TAGS) -o $@ ./cmd/apprun-dedicated-cli

clean:
	rm -rf apprun-dedicated-cli dist/

test:
	go test -tags $(BUILD_TAGS) -v ./...

install:
	go install -tags $(BUILD_TAGS) github.com/fujiwara/apprun-dedicated-cli/cmd/apprun-dedicated-cli

dist:
	goreleaser build --snapshot --clean
