# 可配置变量
BINARY_NAME ?= v2ex_tui
SRC_DIR      ?= .  
OUTPUT_DIR   ?= ./bin
MAIN_PKG     ?= ./cmd/v2ex/main.go
LDFLAGS      ?= -s -w

# 支持的平台组合（格式：os-arch）
ALL_TARGETS = linux-amd64 linux-arm64 \
              darwin-amd64 darwin-arm64 \
              windows-amd64 windows-arm64

.PHONY: build
build:
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) -ldflags "$(LDFLAGS)" $(MAIN_PKG)



	