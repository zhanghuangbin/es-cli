# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目简介

es-cli 是一个基于 Go 的 Elasticsearch 交互式 CLI 工具，用户通过 REPL 模式使用 SQL 语法查询和管理 Elasticsearch。当前处于 MVP 阶段，使用 ES 内置 `_sql` API 执行 SQL，架构上预留了扩展点，后续计划自行实现 SQL 解析器（SQL → ES DSL）。

## 常用命令

```bash
# 构建
go build -o ./_output/es-cli.exe ./cmd/es-cli/

# 运行（连接本地 ES）
go run cmd/es-cli/main.go --address http://localhost:9200

# 运行（带认证）
go run cmd/es-cli/main.go --address https://es-host:9200 --username elastic --password xxx --ca-cert /path/to/ca.crt

# 编译检查
go build ./...

# 依赖整理
go mod tidy

# 运行测试（当前尚无测试文件）
go test ./...
```

## 可复用模块文档

模块详细说明位于 `doc/references` 下。当需要了解具体模块时，请主动读取[summary.md](docs%2Freferences%2Fsummary.md)，并根据其指示，阅读对应的模块的`.md`说明。

重要:

- 永远以**用户的说明**为准
- 新增/修改功能后，主动更新对应的文件说明。
- references的文档说明，可能落后于实际实现，如果发生矛盾，则以实际代码为准

## 项目语言

项目 UI 文本、错误信息、注释均使用中文。代码中的变量名/函数名使用英文。
