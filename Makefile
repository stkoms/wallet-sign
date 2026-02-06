# Lotus Sign Makefile

# 变量定义
APP_NAME := wallet-sign
VERSION := 1.0.0
GO := go
GOFLAGS := -v
LDFLAGS := -s -w -X main.Version=$(VERSION)

# 默认目标
.PHONY: all
all: build

# 编译
.PHONY: build
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(APP_NAME) main.go

# 编译到 build 目录
.PHONY: build-dir
build-dir:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) main.go

# 安装到 GOPATH/bin
.PHONY: install
install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)"

# 运行
.PHONY: run
run:
	$(GO) run main.go

# 测试
.PHONY: test
test:
	$(GO) test -v ./...

# 测试覆盖率
.PHONY: test-coverage
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# 代码检查
.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "请安装 golangci-lint" && exit 1)
	golangci-lint run ./...

# 格式化代码
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# 整理依赖
.PHONY: tidy
tidy:
	$(GO) mod tidy

# 下载依赖
.PHONY: deps
deps:
	$(GO) mod download

# 清理
.PHONY: clean
clean:
	$(GO) clean
	rm -f $(APP_NAME)

	rm -f coverage.out coverage.html

# 跨平台编译 - Linux
.PHONY: build-linux
build-linux: build-linux-amd64 build-linux-arm64 build-linux-386

.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-linux-amd64 main.go

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-linux-arm64 main.go

.PHONY: build-linux-386
build-linux-386:
	GOOS=linux GOARCH=386 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-linux-386 main.go

# 跨平台编译 - macOS
.PHONY: build-darwin
build-darwin: build-darwin-amd64 build-darwin-arm64

.PHONY: build-darwin-amd64
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-darwin-amd64 main.go

.PHONY: build-darwin-arm64
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-darwin-arm64 main.go

# 跨平台编译 - Windows
.PHONY: build-windows
build-windows: build-windows-amd64 build-windows-arm64 build-windows-386

.PHONY: build-windows-amd64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-windows-amd64.exe main.go

.PHONY: build-windows-arm64
build-windows-arm64:
	GOOS=windows GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-windows-arm64.exe main.go

.PHONY: build-windows-386
build-windows-386:
	GOOS=windows GOARCH=386 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-windows-386.exe main.go

# 跨平台编译 - FreeBSD
.PHONY: build-freebsd
build-freebsd:
	GOOS=freebsd GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-freebsd-amd64 main.go
	GOOS=freebsd GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-freebsd-arm64 main.go

# 编译所有平台
.PHONY: build-all
build-all: build-linux build-darwin build-windows build-freebsd


# 帮助信息
.PHONY: help
help:
	@echo "Lotus Sign Makefile"
	@echo ""
	@echo "用法: make [target]"
	@echo ""
	@echo "基础命令:"
	@echo "  build              编译项目"
	@echo "  build-dir          编译到 build 目录"
	@echo "  install            安装到 GOPATH/bin"
	@echo "  run                运行项目"
	@echo "  clean              清理编译产物"
	@echo ""
	@echo "测试与检查:"
	@echo "  test               运行测试"
	@echo "  test-coverage      生成测试覆盖率报告"
	@echo "  lint               代码检查"
	@echo "  fmt                格式化代码"
	@echo ""
	@echo "依赖管理:"
	@echo "  tidy               整理依赖"
	@echo "  deps               下载依赖"
	@echo ""
	@echo "跨平台编译:"
	@echo "  build-linux        编译 Linux 版本 (amd64/arm64/386)"
	@echo "  build-darwin       编译 macOS 版本 (amd64/arm64)"
	@echo "  build-windows      编译 Windows 版本 (amd64/arm64/386)"
	@echo "  build-freebsd      编译 FreeBSD 版本 (amd64/arm64)"
	@echo "  build-all          编译所有平台版本"
	@echo ""
	@echo ""
	@echo "  help               显示帮助信息"
