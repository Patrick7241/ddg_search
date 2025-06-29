# ddg_search

**Language Switch**: [中文](README.md) | [English](README_en.md)

`ddg_search` 是一个用 Go 语言实现的 **非官方 DuckDuckGo 搜索接口封装**，通过模拟 DuckDuckGo 网页请求，提供文本、图片、视频、新闻搜索功能，支持多种搜索模式、区域、时间限制及安全搜索设置。

> **注意**: 该库非 DuckDuckGo 官方 API，可能因网站变动导致失效。

---

## 功能

* 基于 HTTP 请求模拟 DuckDuckGo 搜索
* 支持 `HTML` 和 `Lite` 两种搜索结果后端
* 支持搜索区域、时间范围限制（一天、一周、一月、一年、全部）
* 支持安全搜索级别配置（开、适中、关）
* 支持设置代理、请求超时、请求间隔时间（防止频率限制）
* 支持最大结果数限制
* 提供命令行工具

---

## 安装

```bash
go get github.com/Patrick7241/ddg_search
````

---

## 快速开始

```go
ddgs := ddg_search.NewDDGS(
	ddg_search.WithProxy("your-proxy"),               // 可选，设置代理
	ddg_search.WithTimeout(10*time.Second),           // 可选，设置请求超时
	ddg_search.WithSleepDuration(1500*time.Millisecond), // 可选，设置请求间隔
)

results, err := ddgs.Text(
	"golang",                          // 关键词
	"wt-wt",                           // 地区代码
	ddg_search.SafeSearchModerate,     // 安全搜索等级
	ddg_search.TimelimitAll,           // 时间限制
	ddg_search.BackendLite,            // 搜索路径
	10,                                // 最大结果数
)
if err != nil {
	panic(err)
}

for i, r := range results {
	fmt.Printf("%d. %s\nURL: %s\n摘要: %s\n\n", i+1, r["title"], r["href"], r["body"])
}
```

---

## 命令行工具

本项目提供一个命令行程序，支持直接通过终端进行 DuckDuckGo 搜索。

### 编译

在项目根目录执行：

```bash
go build .\cli\ 
```

将生成一个可执行文件 `cli`。

### 命令行参数

| 参数   | 说明                                         |
| ---- | ------------------------------------------ |
| `-q` | **必填**，搜索关键词                               |
| `-m` | 搜索模式：`text`（默认）、`images`、`news`、`videos`   |
| `-s` | 安全搜索：`on`、`moderate`（默认）、`off`             |
| `-t` | 时间限制：`d`(1天)、`w`(1周)、`m`(1月)、`y`(1年)、空(全部) |
| `-p` | 代理地址（如 `127.0.0.1:7890`），可选                |
| `-n` | 最大结果数，默认 10                                |

### 使用示例

**查看帮助：**
```bash
.\cli.exe -h
```

**文本搜索：**

```bash
.\cli.exe -q "golang" -m text -n 5
```

**图片搜索：**

```bash
.\cli.exe -q "cat" -m images
```

**新闻搜索（过去一周）：**

```bash
.\cli.exe -q "ai news" -m news -t w
```

**视频搜索（使用代理）：**

```bash
.\cli.exe -q "golang tutorial" -m videos -p "127.0.0.1:7890"
```

---

## 主要类型和参数

| 参数名               | 类型     | 说明                                                                                  |
| ----------------- | ------ | ----------------------------------------------------------------------------------- |
| `SafeSearchLevel` | string | 安全搜索等级：`SafeSearchOn`，`SafeSearchModerate`，`SafeSearchOff`                          |
| `Backend`         | string | 路径选择：`BackendAuto`，`BackendHTML`，`BackendLite`                                      |
| `Timelimit`       | string | 时间限制：`TimelimitDay`，`TimelimitWeek`，`TimelimitMonth`，`TimelimitYear`，`TimelimitAll` |

---

## 可选配置函数（Option）

* `WithProxy(proxy string)` 设置 HTTP 代理（如 `127.0.0.1:7890`）
* `WithTimeout(timeout time.Duration)` 设置 HTTP 请求超时，默认 10 秒
* `WithSleepDuration(duration time.Duration)` 设置请求间隔，默认 1500ms，防止频率限制

---

## 错误

* `ErrRatelimit` 请求频率限制错误
* `ErrTimeout` 请求超时
* `ErrSearch` 搜索请求错误
* `ErrInvalidParams` 参数错误

---

## 注意事项

* 该库依赖于网页爬取，DuckDuckGo 网站结构变动可能导致功能失效
* 请合理设置请求间隔和超时，避免被封禁IP
* 本库适合非商业及个人项目使用

---

## 贡献

欢迎提出 issue 或 PR，共同完善功能与稳定性。

---

## License

此项目采用 [MIT License](LICENSE) 许可证，详情请查看 [LICENSE](LICENSE) 文件。

---