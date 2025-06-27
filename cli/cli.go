package main

import (
	"flag"
	"fmt"
	"github.com/Patrick7241/ddg_search"
	"log"
	"strings"
)

var (
	query      string
	mode       string
	safesearch string
	timelimit  string
	proxy      string
	maxResults int
)

func main() {

	flag.StringVar(&query, "q", "", "Search keywords (required)")
	flag.StringVar(&mode, "m", "text", "Search mode: text | images | news | videos")
	flag.StringVar(&safesearch, "s", "moderate", "Safe search: on | moderate | off")
	flag.StringVar(&timelimit, "t", "", "Time limit: d | w | m | y")
	flag.StringVar(&proxy, "p", "", "Proxy address (e.g., 127.0.0.1:7890)")
	flag.IntVar(&maxResults, "n", 10, "Max number of results")

	flag.Parse()

	if query == "" {
		log.Fatal("Please provide search keywords using -q")
	}

	var client *ddg_search.DDGS
	if proxy != "" {
		client = ddg_search.NewDDGS(
			ddg_search.WithProxy(proxy),
		)
	} else {
		client = ddg_search.NewDDGS()
	}

	// SafeSearch enum
	var safe ddg_search.SafeSearchLevel
	switch strings.ToLower(safesearch) {
	case "on":
		safe = ddg_search.SafeSearchOn
	case "moderate":
		safe = ddg_search.SafeSearchModerate
	case "off":
		safe = ddg_search.SafeSearchOff
	default:
		safe = ddg_search.SafeSearchModerate
	}

	// Timelimit enum
	var time ddg_search.Timelimit
	switch timelimit {
	case "d":
		time = ddg_search.TimelimitDay
	case "w":
		time = ddg_search.TimelimitWeek
	case "m":
		time = ddg_search.TimelimitMonth
	case "y":
		time = ddg_search.TimelimitYear
	default:
		time = ddg_search.TimelimitAll
	}

	switch mode {
	case "text":
		results, err := client.Text(query, "wt-wt", safe, time, ddg_search.BackendAuto, maxResults)
		if err != nil {
			log.Fatal("Search error:", err)
		}
		for i, r := range results {
			fmt.Printf("[%d] title: %s\n href: %s\n body: %s\n\n", i+1, r["title"], r["href"], r["body"])
		}
	case "images":
		results, err := client.Images(query, "wt-wt", safe, time, maxResults)
		if err != nil {
			log.Fatal("Search error:", err)
		}
		for i, r := range results {
			fmt.Printf("[%d] image: %s\n\n", i+1, r["image"])
		}
	case "news":
		results, err := client.News(query, "wt-wt", safe, time, maxResults)
		if err != nil {
			log.Fatal("Search error:", err)
		}
		for i, r := range results {
			fmt.Printf("[%d] title: %s\n url: %s\n body: %s\n\n", i+1, r["title"], r["url"], r["body"])
		}
	case "videos":
		results, err := client.Videos(
			query,
			"wt-wt",
			safe,
			time,
			ddg_search.ResolutionAll,
			ddg_search.DurationAll,
			ddg_search.LicenseAll,
			maxResults,
		)
		if err != nil {
			log.Fatal("Search error:", err)
		}
		for i, r := range results {
			fmt.Printf("[%d] title: %s\n content: %s\n\n", i+1, r["title"], r["content"])
		}
	default:
		log.Fatalf("Unknown mode: %s", mode)
	}
}
