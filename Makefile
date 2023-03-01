BUILD_DIR=./cmd/...

.PHONY: build test release.arm64 release.armv7 release.amd64

build:
	go mod verify
	go build -o dist/ $(BUILD_FLAGS) $(BUILD_DIR)

release.arm64:
	GOOS=linux GOARCH=arm64 go build -o dist/linux-arm64/ $(BUILD_FLAGS) $(BUILD_DIR)
	
release.armv7:
	GOOS=linux GOARCH=arm GOARM=7 go build -o dist/linux-armv7/ $(BUILD_FLAGS) $(BUILD_DIR)
	
release.amd64:
	GOOS=linux GOARCH=amd64 go build -o dist/linux-amd64/ $(BUILD_FLAGS) $(BUILD_DIR)
	

release: release.arm64 release.armv7 release.amd64

test:
	go test -v -race ./...
