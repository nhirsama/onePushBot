# onePushBot

> 一个用于通过调用 QQ API 接口实现某些自动化功能的小项目。

---

## 功能简介

onePushBot 是一个用 Go 编写的轻量级机器人框架，通过调用 [NapCat](https://napneko.github.io/) 的 API / WebSocket 接口来实现自动化功能（消息处理、事件监听、插件式扩展等）。

## 主要特性

* 使用 Go 语言实现，结构清晰、易于编译和部署。
* 基于 [NapCat](https://napneko.github.io/) 的 WebSocket 接口与 QQ API 交互。
* 可以并发处理通过 WebSocket 接收到的每一条信息，并保证 API 接口的并发安全。

## 开发中特性

* Web 管理界面
* 插件系统
* 仍有 NapCat API 接口未实现

## 环境要求

* Go 1.18+（建议使用最新稳定版）

## 快速开始（本地运行）

```bash
# 克隆仓库
git clone https://github.com/nhirsama/onePushBot.git
cd onePushBot

# 拉取依赖并构建
go mod download
go build -o onePushBot ./

# 运行（如果有 config 文件或环境变量，先准备好）
./onePushBot
# 或者
go run ./
```

## 贡献

欢迎提交 issue/PR：

1. Fork 本仓库
2. 新建分支（`feature/xxx` 或 `fix/xxx`）
3. 提交代码并发起 PR，说明修改内容与测试方法

## 许可证

本项目基于 Apache License 2.0 许可证发布，详情请参阅 [LICENSE](./LICENSE) 文件。
