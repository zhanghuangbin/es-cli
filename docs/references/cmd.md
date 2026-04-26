# cmd 模块

> `cmd/es-cli/main.go` + `internal/cmd/`

程序入口。`main.go` 仅调用 `cmd.Execute()`。

## 文件结构

| 文件 | 说明 |
|------|------|
| `root.go` | 定义根命令（进入 REPL 模式）、全局 CLI 参数 |
| `client.go` | `newESClient()` 公共函数，创建 ES 客户端并 ping 验证 |
| `version.go` | `version` 子命令，打印版本信息 |
| `exec.go` | `exec` 子命令，执行单条 SQL 并格式化输出 |

## 全局 CLI 参数

`--address`（ES 地址）、`--username`、`--password`、`--password-stdin`、`--ca-cert`

## 子命令

### version

打印版本、Git 提交、构建日期、Go 版本。支持通过 `-ldflags` 注入编译信息。

### exec

执行单条 SQL 并以指定格式输出结果，适用于脚本化和管道化使用。

| Flag | 短写 | 默认值 | 说明 |
|------|------|--------|------|
| `--command` | `-c` | — | 要执行的 SQL 语句（必填） |
| `--format` | `-f` | `json` | 输出格式（json, csv, yaml, go-template） |
| `--jsonpath` | — | — | JSONPath 表达式（仅 json 格式有效） |
| `--template` | — | — | Go 模板字符串（仅 go-template 格式有效） |
| `--field` | — | — | 输出的列名（仅 csv 格式有效，可多次指定） |

## 构建

```bash
# 普通构建
go build -o ./_output/es-cli.exe ./cmd/es-cli/

# 带版本信息构建
go build -ldflags "-X github.com/zhanghuangbin/es-cli/internal/version.Version=v0.1.0 \
  -X github.com/zhanghuangbin/es-cli/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X github.com/zhanghuangbin/es-cli/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o ./_output/es-cli.exe ./cmd/es-cli/
```
