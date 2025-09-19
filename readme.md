# Agent-Go 聚合搜索演示服务

该项目实现了一个 Go 语言编写的聚合搜索服务雏形，用于模拟“知乎 / 微信公众号 / 小红书”等平台的最新内容查询，并在服务端自动做结构化总结。虽然目前内置的爬取源为本地模拟数据，但整体架构（Provider 接口 + 聚合器 + 摘要器 + HTTP API）已经按真实项目拆分，后续只需替换或扩展 Provider 模块即可接入真实数据源。

## 功能特性

- 🌐 **多源聚合**：通过 `Provider` 接口统一各平台的搜索能力，支持并发请求、去重、排序与缓存。
- 🧠 **自动摘要**：内置 `Simple` 摘要器，会根据结果生成要点、关键词、来源分布与情绪倾向。
- 📝 **历史记录**：记录最近若干次查询，便于运营和分析人员回顾。
- ⚙️ **可配置化**：支持通过环境变量调整端口、缓存 TTL、超时时间及启用的 Provider。

## 目录结构

```
cmd/server/           # 可执行程序入口
internal/aggregator/  # 聚合逻辑、缓存调度
internal/provider/    # 数据源 Provider，当前内置 mock 数据
internal/summary/     # 摘要与分析
internal/httpserver/  # HTTP 接口封装（net/http）
internal/config/      # 环境变量解析
internal/cache/       # 简单内存缓存
internal/history/     # 查询历史存储
```

## 快速开始

1. **安装依赖**（Go 1.21+）：
   ```bash
   go mod tidy
   ```
2. **运行服务**：
   ```bash
   go run ./cmd/server
   ```
   默认监听 `:8080`，启动后可通过以下接口测试：
   - `GET /healthz`：存活检测
   - `GET /v1/providers`：列出可用 Provider
   - `GET /v1/search?q=运营`：执行查询并返回聚合结果与自动摘要
   - `GET /v1/history`：查看最近的查询记录

3. **调整配置**（示例）：
   ```bash
   export APP_PORT=8090
   export CACHE_TTL=2m
   export REQUEST_TIMEOUT=5s
   export PROVIDERS=mock
   go run ./cmd/server
   ```

> ⚠️ 说明：当前 `mock` Provider 使用预置的示例数据，便于在无外网或未取得平台授权的情况下演示流程。若要接入真实数据，只需在 `internal/provider` 下实现新的 Provider 并在 `cmd/server/main.go` 注册即可。

## 测试

```bash
go test ./...
go vet ./...
```

## 下一步可扩展方向

- 接入真实抓取逻辑（配合代理、验证码处理、频率限制）。
- 引入消息队列与异步任务，处理大规模抓取与缓存刷新。
- 增强摘要模块（引入大语言模型、情绪分析或聚类）。
- 为管理后台和订阅推送预留接口。
