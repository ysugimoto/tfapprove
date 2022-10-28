.PHONY: build darwin_arm64 darwin_amd64 linux windows_amd64

BUILD_VERSION=$(or ${VERSION}, dev)
AGGREGATE_SERVER=$(or ${SERVER}, "")

darwin_arm64:
	GOOS=darwin GOARCH=arm64 go build \
			 -ldflags "-X main.aggregateServer=$(AGGREGATE_SERVER)" \
			 -ldflags "-X main.version=$(BUILD_VERSION)" \
			 -o dist/tfapprove-darwin-arm64 .

darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build \
			 -ldflags "-X main.aggregateServer=$(AGGREGATE_SERVER)" \
			 -ldflags "-X main.version=$(BUILD_VERSION)" \
			 -o dist/tfapprove-darwin-amd64 .

linux:
	GOOS=linux GOARCH=amd64 go build \
			 -ldflags "-X main.aggregateServer=$(AGGREGATE_SERVER)" \
			 -ldflags "-X main.version=$(BUILD_VERSION)" \
			 -o dist/tfapprove-linux-amd64 .

windows_amd64:
	GOOS=windows GOARCH=amd64 go build \
			 -ldflags "-X main.aggregateServer=$(AGGREGATE_SERVER)" \
			 -ldflags "-X main.version=$(BUILD_VERSION)" \
			 -o dist/tfapprove-windows-amd64.exe .

build: linux darwin_amd64 darwin_arm64 windows_amd64

