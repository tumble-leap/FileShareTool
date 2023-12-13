GOBIN = go

BUILD_DIR = build
BINARY_NAME = file-share

ENV = GOPROXY=https://goproxy.cn,direct

build-frontend:
	cd webapp && npm install
	cd webapp && npm run build

build:
	$(ENV) $(GOBIN) build -o $(BUILD_DIR)/$(BINARY_NAME)

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(ENV) $(GOBIN) build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(ENV) $(GOBIN) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(ENV) $(GOBIN) build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64

build-all: build-windows build-linux build-darwin

clean:
	rm -rf $(BUILD_DIR)
	rm -rf webapp/$(BUILD_DIR)

.PHONY: build-frontend build build-windows build-linux build-darwin build-all clean
