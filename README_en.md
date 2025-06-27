# ddg\_search

**Language Switch**: [中文](README.md) | [English](README_en.md)

`ddg_search` is an **unofficial DuckDuckGo search API wrapper** implemented in Go. It simulates DuckDuckGo web requests to provide text, images, news and videos search functionality, supporting multiple search modes, regions, time filters, and safe search settings.

> **Note**: This library is not an official DuckDuckGo API and may break if DuckDuckGo changes their website.

---

## Features

* Simulates HTTP requests to DuckDuckGo search engine
* Supports `HTML` and `Lite` search result backends
* Supports region and time-range filtering (day, week, month, year, all)
* Supports safe search levels (on, moderate, off)
* Supports proxy settings, request timeout, and configurable delay between requests (to avoid rate limiting)
* Supports maximum results limit

---

## Installation

```bash
go get github.com/Patrick7241/ddg_search
```

---

## Quick Start

``` go
	ddgs := ddg_search.NewDDGS(
		ddg_search.WithProxy("your-proxy"),                // optional proxy
		ddg_search.WithTimeout(10*time.Second),           // optional timeout
		ddg_search.WithSleepDuration(1500*time.Millisecond), // optional request delay
	)

	results, err := ddgs.Text(
		"golang",                        // keywords
		"wt-wt",                        // region code, default "wt-wt"
		ddg_search.SafeSearchModerate,  // safe search level
		ddg_search.TimelimitAll,        // time filter
		ddg_search.BackendLite,         // backend type (Auto, HTML, Lite)
		10,                            // max results
	)
	if err != nil {
		panic(err)
	}

	for i, r := range results {
		fmt.Printf("%d. %s\nURL: %s\nSnippet: %s\n\n", i+1, r["title"], r["href"], r["body"])
	}
```

---

## Main Types and Parameters

| Parameter         | Type   | Description                                                                                      |
| ----------------- | ------ | ------------------------------------------------------------------------------------------------ |
| `SafeSearchLevel` | string | Safe search levels: `SafeSearchOn`, `SafeSearchModerate`, `SafeSearchOff`                        |
| `Backend`         | string | Backend options: `BackendAuto`, `BackendHTML`, `BackendLite`                                     |
| `Timelimit`       | string | Time filters: `TimelimitDay`, `TimelimitWeek`, `TimelimitMonth`, `TimelimitYear`, `TimelimitAll` |

---

## Optional Configuration Functions (Options)

* `WithProxy(proxy string)` — Set HTTP proxy (e.g. `127.0.0.1:7890`)
* `WithTimeout(timeout time.Duration)` — Set HTTP request timeout, default 10s
* `WithSleepDuration(duration time.Duration)` — Set delay between requests, default 1500ms (to avoid rate limiting)

---

## Errors

* `ErrRatelimit` — Rate limiting error
* `ErrTimeout` — Request timeout error
* `ErrSearch` — Search request error
* `ErrInvalidParams` — Invalid parameter error

---

## Notes

* This library relies on web scraping and may break if DuckDuckGo website structure changes
* Please set reasonable request delays and timeouts to avoid IP bans
* Intended for personal and non-commercial use

---

## Contribution

Issues and pull requests are welcome to improve features and stability.

---