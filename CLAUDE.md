# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 构建与测试

```bash
go build .                          # 编译二进制文件
go test ./...                       # 运行所有测试
go test -run TestBuild ./internal/message/  # 运行单个测试
```

## 架构

三个子命令，各自构成数据处理流水线的一个阶段：

- **`collect`** — 将 `kw_city.txt × kw_hospital.txt` 做笛卡尔积生成搜索关键词，使用 10 个并发 worker、随机延迟和轮换 User-Agent 在百度搜索。用正则提取广告落地页 URL（`ada.baidu.com/site/.../xyl?imid=...`），去重后写入 `api.txt`。支持可选 HTTP 代理。

- **`send`** — 从 `api.txt` 读取 URL，打乱顺序，通过 [go-rod](https://github.com/go-rod/rod) 启动可见的 Chrome 实例。对每个 URL，导航到页面，定位输入框（`.imlp-component-typebox-input`）和发送按钮（`.imlp-component-typebox-send-btn`），填入随机组合的留言并点击发送。

- **`cron`** — 通过 [robfig/cron](https://github.com/robfig/cron/v3) 将 `send` 包装为每天早上 9:00 的定时任务，支持 Ctrl+C 优雅退出。

## 模块关系

```
main.go                     — CLI 分发（collect|send|cron），加载 config.yaml
internal/config/config.go   — YAML 配置结构体与 Load()
internal/collector/         — 百度搜索 + 正则提取广告 URL
internal/sender/            — 基于 Rod 的浏览器自动化表单提交
internal/message/builder.go — 从配置模板随机组合留言
```

数据流向：`data/*.txt` → collector → `api.txt` → sender → 浏览器表单提交

## 核心依赖

| 包 | 用途 |
|---|---|
| `github.com/go-rod/rod` | Chrome DevTools Protocol 浏览器自动化 |
| `github.com/robfig/cron/v3` | 定时任务调度 |
| `gopkg.in/yaml.v3` | YAML 配置文件解析 |

## 配置

`config.yaml` 同时驱动搜索短语模板（titles、relatives、situations、contact_methods、greetings）和运行时设置（proxy、browser_path、tel_number、tel_name）。留言构建器从每个类别中随机选取一项，并将 `{number}` 替换为配置的电话号码。
