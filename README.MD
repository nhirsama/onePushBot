# onePushBot

> 一个用于通过调用 QQ API 接口实现某些自动化功能的小项目。

---

## 功能简介

onePushBot 是一个用 Go 编写的轻量级机器人框架，通过调用 [NapCat](https://napneko.github.io/) 的 API / WebSocket
接口来实现自动化功能（消息处理、事件监听、插件式扩展等）。

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

## 文档

### 配置约定

项目只保留一套配置键名，不再兼容历史别名。运行时配置保存在 SQLite `settings` 表中，管理面板也只读写这些规范键。

常用配置分组：

| 分组                          | 说明                                                  |
|:----------------------------|:----------------------------------------------------|
| `platforms`                 | 已启用平台列表，例如 `["qq","telegram_bot"]`                  |
| `modules`                   | 已启用模块列表，例如 `["reply","sudo","riddle"]`              |
| `qq.*`                      | NapCat / QQ 平台配置                                    |
| `telegram_bot.*`            | Telegram Bot 配置                                     |
| `telegram_user.*`           | Telegram 用户态配置                                      |
| `feishu.*`                  | 飞书平台配置                                              |
| `llm.*`                     | 大模型配置，例如 `llm.api_token`、`llm.model`、`llm.base_url` |
| `reply.*`                   | `reply` 模块配置                                        |
| `sudo.*`                    | `sudo` 模块配置                                         |
| `riddle.*`                  | `riddle` 模块配置                                       |
| `auto_set_msg_emoji_like.*` | 自动表情回应模块配置                                          |
| `group_message_token_db.*`  | 群消息分词统计模块配置                                         |
| `nowcoder_daily.*`          | 牛客每日一题推送配置                                          |

### 事件订阅字段

业务路由通过 `platform.EventFilter` 订阅总线事件。字段为空时表示不限制该条件。

| 字段 | 类型 | 格式 / 取值 | 说明 |
|:--|:--|:--|:--|
| `Platform` | `platform.Platform` | `qq`、`telegram_user`、`telegram_bot`、`feishu` | 来源平台。 |
| `Kind` | `platform.EventKind` | `message`、`notice`、`request` | 事件大类。 |
| `SubType` | `string` | 平台内事件子类型 | 例如 QQ `group`、`private`、`poke`、`friend`、`group`，Telegram `group_message`、`private_message`、`channel_post`。 |
| `ChatType` | `platform.ChatType` | `private`、`group`、`channel`、`unknown` | 会话类型。 |
| `ChatID` | `string` | 平台原始 ID 的字符串形式 | 群号、用户会话 ID、频道 ID、飞书 chat_id 等。 |
| `UserID` | `string` | 平台原始 ID 的字符串形式 | 消息发送者、通知触发用户、请求发起用户。 |

常用订阅：

| 需求 | 订阅条件 |
|:--|:--|
| 所有群聊消息 | `EventFilter{Kind: EventKindMessage, ChatType: ChatTypeGroup}` |
| 所有私聊消息 | `EventFilter{Kind: EventKindMessage, ChatType: ChatTypePrivate}` |
| 指定 QQ 群消息 | `EventFilter{Platform: PlatformQQ, Kind: EventKindMessage, ChatType: ChatTypeGroup, ChatID: "123456"}` |
| QQ 拍一拍 | `EventFilter{Platform: PlatformQQ, Kind: EventKindNotice, SubType: "poke"}` |
| QQ 好友/群请求 | `EventFilter{Platform: PlatformQQ, Kind: EventKindRequest}` |
| 飞书消息 | `EventFilter{Platform: PlatformFeishu, Kind: EventKindMessage}` |
| Telegram 用户态频道消息 | `EventFilter{Platform: PlatformTelegramUser, Kind: EventKindMessage, ChatType: ChatTypeChannel}` |

### 事件结构

订阅命中后，处理器收到 `platform.Event`。

| 字段 | 类型 | 说明 |
|:--|:--|:--|
| `ID` | `string` | 事件 ID。消息事件通常等于消息 ID；通知/请求事件由平台字段组合生成。 |
| `Platform` | `platform.Platform` | 事件来源平台。 |
| `Kind` | `platform.EventKind` | `message`、`notice`、`request`。 |
| `SubType` | `string` | 平台内子类型，用于进一步区分同一大类事件。 |
| `Time` | `time.Time` | 事件发生时间。平台未提供时使用接收时间。 |
| `SelfID` | `string` | 收到该事件的平台账号 ID。平台无法提供时为空。 |
| `Message` | `*platform.Message` | `Kind == message` 时非空。 |
| `Notice` | `*platform.Notice` | `Kind == notice` 时非空。 |
| `Request` | `*platform.Request` | `Kind == request` 时非空。 |
| `Raw` | `any` | 平台原始事件对象，平台专属逻辑兜底使用。 |

### 消息事件

`Event.Kind == EventKindMessage` 时读取 `Event.Message`。

| 字段 | 类型 | 说明 |
|:--|:--|:--|
| `ID` | `string` | 消息 ID。 |
| `Chat` | `platform.Chat` | 会话信息。 |
| `Sender` | `platform.User` | 发送者。 |
| `Text` | `string` | 提取后的文本内容。非文本消息可能为空。 |
| `RawText` | `string` | 平台原始文本。QQ 对应 `raw_message`。 |
| `Segments` | `[]platform.Segment` | 消息段列表。QQ 文本、at、图片等会保留在这里。 |
| `Time` | `time.Time` | 消息时间。 |
| `Target` | `platform.User` | 目标用户。仅部分平台/事件提供。 |
| `Font` | `int` | QQ 字体字段，其他平台通常为 `0`。 |
| `SentBySelf` | `bool` | 是否为当前账号自己发出的消息。QQ `message_sent` 为 `true`。 |
| `DetailType` | `string` | 平台消息细分类型。QQ 通常是 `normal`、`friend` 等；飞书通常是 `text`、`image` 等。 |
| `Anonymous` | `any` | QQ 匿名群消息信息。没有则为空。 |
| `PlatformData` | `any` | 平台消息原始对象。 |

### 通知事件

`Event.Kind == EventKindNotice` 时读取 `Event.Notice`。

| 字段 | 类型 | 说明 |
|:--|:--|:--|
| `Type` | `string` | 通知类型。QQ 例如 `group_upload`、`group_admin`、`group_decrease`、`group_increase`、`group_ban`、`friend_add`、`group_recall`、`friend_recall`、`poke`。 |
| `Chat` | `platform.Chat` | 关联会话。没有会话信息时字段为空。 |
| `User` | `platform.User` | 通知触发用户。 |
| `Target` | `platform.User` | 被操作目标用户。 |
| `Operator` | `platform.User` | 操作者。QQ 管理、禁言、文件等通知可能提供。 |
| `MessageID` | `string` | 关联消息 ID。撤回、已读等通知可能提供。 |
| `FileID` | `string` | 文件 ID。文件上传通知可能提供。 |
| `FileName` | `string` | 文件名。文件上传通知可能提供。 |
| `FileSize` | `int64` | 文件大小，单位字节。 |
| `Duration` | `time.Duration` | 持续时间。QQ 禁言通知会使用该字段。 |
| `DetailType` | `string` | 平台通知细分类型。QQ `notify.poke` 中 `Type` 为 `poke`，`DetailType` 为 `notify`。 |
| `PlatformData` | `any` | 平台通知原始对象。 |

### 请求事件

`Event.Kind == EventKindRequest` 时读取 `Event.Request`。

| 字段 | 类型 | 说明 |
|:--|:--|:--|
| `Type` | `string` | 请求类型。QQ 常见为 `friend`、`group`。 |
| `Chat` | `platform.Chat` | 关联群聊。好友请求通常为空。 |
| `User` | `platform.User` | 请求发起用户。 |
| `Comment` | `string` | 验证信息。 |
| `Flag` | `string` | 平台处理请求需要用到的标识。 |
| `DetailType` | `string` | 请求子类型。QQ 群请求常见为 `add`、`invite`。 |
| `PlatformData` | `any` | 平台请求原始对象。 |

### 公共子结构

| 类型 | 字段 | 说明 |
|:--|:--|:--|
| `Chat` | `ID string` | 会话 ID。 |
| `Chat` | `Type ChatType` | `private`、`group`、`channel`、`unknown`。 |
| `Chat` | `Name string` | 群名、频道名等。平台未提供时为空。 |
| `User` | `ID string` | 用户 ID。 |
| `User` | `Name string` | 优先选择出的显示名。 |
| `User` | `Nickname/Card/Remark/Role/Title/Level string` | 平台能提供时填充；QQ sender 中常见。 |
| `Segment` | `Type string` | 消息段类型，例如 `text`、`at`、`image`。 |
| `Segment` | `Text string` | 可直接拼接的文本。非文本段可能为空。 |
| `Segment` | `Data any` | 平台消息段原始数据。 |
