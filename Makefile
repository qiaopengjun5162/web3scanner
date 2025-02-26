.DEFAULT_GOAL := web3scanner

# 获取 Git 提交哈希和时间，默认值处理非 Git 环境
GITCOMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GITDATE := $(shell git show -s --format='%ct' 2>/dev/null || echo "0")

# 构造链接器标志
LDFLAGS := -ldflags "-X main.GitCommit=$(GITCOMMIT) -X main.GitDate=$(GITDATE)"
PROJECT_NAME := $(shell go list -m | awk -F/ '{print $$NF}')

# 整理 Go 模块依赖
tidy:
	go mod tidy

# 编译 web3scanner 程序，嵌入 Git 提交信息
web3scanner: tidy
	go build -v $(LDFLAGS) -o $(PROJECT_NAME) ./cmd/web3scanner

# 清理生成的文件和 Go 缓存
clean:
	rm -f web3scanner
	go clean -cache -testcache

# 运行所有测试
test: tidy
	go test -v ./...

# 检查代码风格和潜在问题
lint: tidy
	golangci-lint run ./...

# 编译协议文件
proto:
	@test -f ./bin/compile.sh || (echo "compile.sh not found" && exit 1)
	sh ./bin/compile.sh

.PHONY: web3scanner clean test lint proto tidy
