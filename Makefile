.PHONY: help build run test clean install-deps format lint

# 变量定义
BINARY_NAME=mini-tmk-agent
MAIN_PATH=./cmd/mini-tmk-agent
GO=go
GOFLAGS=-v

help:
	@echo "Mini TMK Agent - 开发命令"
	@echo ""
	@echo "可用命令:"
	@echo "  make build           - 编译项目"
	@echo "  make run-stream      - 运行 Stream 模式"
	@echo "  make run-transcript  - 运行 Transcript 模式"
	@echo "  make test            - 运行测试"
	@echo "  make clean           - 清理编译产物"
	@echo "  make install-deps    - 安装依赖"
	@echo "  make format          - 格式化代码"
	@echo "  make lint            - 代码检查"
	@echo "  make help            - 显示此帮助信息"

build:
	@echo "正在编译 $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "编译完成: ./$(BINARY_NAME)"

run-stream: build
	@echo "启动 Stream 模式..."
	./$(BINARY_NAME) stream --source-lang zh --target-lang en --verbose

run-transcript: build
	@echo "启动 Transcript 模式..."
	@if [ ! -f "sample.mp3" ]; then \
		echo "错误: 未找到 sample.mp3 文件"; \
		exit 1; \
	fi
	./$(BINARY_NAME) transcript --file sample.mp3 --output output.txt --source-lang zh --target-lang en --verbose

test:
	@echo "运行测试..."
	$(GO) test -v ./...

clean:
	@echo "清理编译产物..."
	$(GO) clean
	rm -f $(BINARY_NAME)
	@echo "清理完成"

install-deps:
	@echo "下载依赖..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "依赖安装完成"

format:
	@echo "格式化代码..."
	$(GO) fmt ./...
	@echo "格式化完成"

lint:
	@echo "代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "未安装 golangci-lint, 请运行: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

dev: clean install-deps format build
	@echo "开发环境准备完成"

.DEFAULT_GOAL := help
