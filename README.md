# es-cli

基于 SQL 的 Elasticsearch 交互式命令行工具。通过 REPL 模式，使用 SQL 语法查询和管理 Elasticsearch。

## 功能特性

- **SQL 查询**: 使用标准 SQL 语法查询 Elasticsearch（基于 ES 内置 `_sql` API）
- **DDL 支持**: `CREATE TABLE`、`DROP TABLE`、`ALTER INDEX/TABLE` 等索引管理操作
- **DML 支持**: `INSERT INTO`、`UPDATE`、`DELETE FROM` 等数据操作
- **智能补全**: SQL 关键词、索引名称自动补全
- **多输出格式**: 支持 table 格式化输出（json/csv 规划中）
- **TLS 认证**: 支持用户名/密码认证和 CA 证书
- **内置命令**: `.help`、`.ping`、`.indices`、`.schema`、`.format` 等快捷操作
- **多行输入**: 支持反斜杠 `\` 续行
- **历史记录**: 支持命令历史回溯

## 安装

### 从源码构建

```bash
# 需要 Go 1.23+
git clone https://github.com/zhanghuangbin/es-cli.git
cd es-cli
go build -o es-cli ./cmd/es-cli/
```

## 快速开始

```bash
# 连接本地 ES
./es-cli --address http://localhost:9200

# 带认证连接
./es-cli --address https://es-host:9200 --username elastic --password xxx

# 使用 CA 证书
./es-cli --address https://es-host:9200 --username elastic --password xxx --ca-cert /path/to/ca.crt

# 从标准输入读取密码
echo "password" | ./es-cli --address https://es-host:9200 --username elastic --password-stdin
```

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--address` | `http://localhost:9200` | Elasticsearch 地址（支持多个） |
| `--username` | - | 用户名 |
| `--password` | - | 密码 |
| `--password-stdin` | `false` | 从标准输入读取密码 |
| `--ca-cert` | - | TLS CA 证书路径 |

## 使用示例

### SQL 查询

```sql
SELECT * FROM my_index LIMIT 10
SELECT name, age FROM users WHERE age > 18 ORDER BY age
SHOW TABLES
DESCRIBE my_index
```

### DDL 操作

```sql
CREATE TABLE my_index (name TEXT, age INTEGER) SETTINGS (number_of_shards=3)
DROP TABLE my_index
ALTER INDEX my_index SETTINGS (number_of_replicas=2)
ALTER TABLE my_index RENAME TO new_index
```

### DML 操作

```sql
INSERT INTO my_index (name, age) VALUES ('张三', 25)
UPDATE my_index SET age=26 WHERE name='张三'
DELETE FROM my_index WHERE name='张三'
```

### 内置命令

```
.help            显示帮助信息
.ping            测试 ES 连接是否正常
.format <类型>   设置输出格式 (table, json*, csv*)
.indices         列出所有索引
.schema <索引名> 显示索引 mapping
.exit            退出 es-cli
Ctrl+D           退出 es-cli
```

## 项目结构

```
es-cli/
├── cmd/es-cli/          # 程序入口
├── internal/
│   ├── cmd/             # CLI 命令定义（cobra）
│   ├── executor/        # SQL 执行器，按类型分发到对应 Handler
│   ├── formatter/       # 输出格式化（table）
│   ├── repl/            # 交互式 REPL（补全、历史、内置命令）
│   └── translator/      # SQL 翻译层（当前使用 ES _sql API）
├── pkg/es/              # Elasticsearch 客户端封装
└── docs/                # 设计文档与参考资料
```

## 技术栈

- [Go](https://go.dev/) 1.23+
- [go-elasticsearch](https://github.com/elastic/go-elasticsearch) - Elasticsearch 官方 Go 客户端
- [cobra](https://github.com/spf13/cobra) - CLI 框架
- [go-prompt](https://github.com/c-bata/go-prompt) - 交互式提示与自动补全
- [go-pretty](https://github.com/jedib0t/go-pretty) - 表格格式化输出

## 许可证

本项目采用 BSD 3-Clause 许可证，详见 [LICENSE](LICENSE) 文件。
