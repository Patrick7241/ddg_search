# ddg\_search

**Language Switch**: [中文](README.md) | [English](README_en.md)

`ddg_search` is an **unofficial DuckDuckGo search API wrapper implemented in Go**. It simulates DuckDuckGo web requests to provide text, image, video, and news search capabilities, supporting multiple search modes, regions, time restrictions, and safe search settings.

> **Note**: This library is not an official DuckDuckGo API and may break if the website changes.

---

## Features

* Simulates DuckDuckGo search via HTTP requests
* Supports `HTML` and `Lite` backends for search results
* Supports region and time restrictions (day, week, month, year, all)
* Configurable safe search levels (on, moderate, off)
* Proxy, timeout, and request interval configuration to avoid rate limits
* Result count limitation
* Includes a command-line tool

---

## Installation

```bash
go get github.com/Patrick7241/ddg_search
```

---

## Quick Start

```go
ddgs := ddg_search.NewDDGS(
	ddg_search.WithProxy("your-proxy"),               // Optional: set proxy
	ddg_search.WithTimeout(10*time.Second),           // Optional: set request timeout
	ddg_search.WithSleepDuration(1500*time.Millisecond), // Optional: set request interval
)

results, err := ddgs.Text(
	"golang",                          // Keywords
	"wt-wt",                           // Region code
	ddg_search.SafeSearchModerate,     // Safe search level
	ddg_search.TimelimitAll,           // Time limit
	ddg_search.BackendLite,            // Backend
	10,                                // Max results
)
if err != nil {
	panic(err)
}

for i, r := range results {
	fmt.Printf("%d. %s\nURL: %s\nSnippet: %s\n\n", i+1, r["title"], r["href"], r["body"])
}
```

---

## Command-Line Tool

This project provides a command-line program for searching DuckDuckGo directly from your terminal.

### Build

Run in the project root directory:

```bash
go build .\cli\
```

This will generate an executable `cli`.

### Command-Line Arguments

| Parameter | Description                                                                 |
| --------- | --------------------------------------------------------------------------- |
| `-q`      | **Required**. Search keywords                                               |
| `-m`      | Search mode: `text` (default), `images`, `news`, `videos`                   |
| `-s`      | Safe search: `on`, `moderate` (default), `off`                              |
| `-t`      | Time limit: `d`(1 day), `w`(1 week), `m`(1 month), `y`(1 year), empty (all) |
| `-p`      | Proxy address (e.g., `127.0.0.1:7890`), optional                            |
| `-n`      | Max number of results (default: 10)                                         |

### Usage Examples

**View help:**

```bash
.\cli.exe -h
```

**Text search:**

```bash
.\cli.exe -q "golang" -m text -n 5
```

**Image search:**

```bash
.\cli.exe -q "cat" -m images
```

**News search (past week):**

```bash
.\cli.exe -q "ai news" -m news -t w
```

**Video search with proxy:**

```bash
.\cli.exe -q "golang tutorial" -m videos -p "127.0.0.1:7890"
```

---

## Key Types and Parameters

| Name              | Type   | Description                                                                                     |
| ----------------- | ------ | ----------------------------------------------------------------------------------------------- |
| `SafeSearchLevel` | string | Safe search levels: `SafeSearchOn`, `SafeSearchModerate`, `SafeSearchOff`                       |
| `Backend`         | string | Backend options: `BackendAuto`, `BackendHTML`, `BackendLite`                                    |
| `Timelimit`       | string | Time limits: `TimelimitDay`, `TimelimitWeek`, `TimelimitMonth`, `TimelimitYear`, `TimelimitAll` |

---

## Optional Configuration Functions

* `WithProxy(proxy string)` Set HTTP proxy (e.g., `127.0.0.1:7890`)
* `WithTimeout(timeout time.Duration)` Set HTTP request timeout (default: 10 seconds)
* `WithSleepDuration(duration time.Duration)` Set request interval (default: 1500ms, to avoid rate limits)

---

## Errors

* `ErrRatelimit` Request rate limit error
* `ErrTimeout` Request timeout error
* `ErrSearch` Search request error
* `ErrInvalidParams` Invalid parameter error

---

## Notes

* This library relies on scraping. If DuckDuckGo changes their website structure, it may stop working.
* Adjust timeouts and intervals to avoid getting your IP banned.
* Intended for non-commercial and personal use.

---

## Contributing

Issues and pull requests are welcome to improve features and stability.

---

## License

This project is licensed under the [MIT License](LICENSE). See the [LICENSE](LICENSE) file for details.

---
