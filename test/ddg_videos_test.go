package test

import (
	"fmt"
	"github.com/Patrick7241/ddg_search"
	"testing"
	"time"
)

func TestVideosDDG(t *testing.T) {
	ddgs := ddg_search.NewDDGS(
		ddg_search.WithProxy("127.0.0.1:7890"), // add proxy
		ddg_search.WithTimeout(10*time.Second),
		ddg_search.WithSleepDuration(10*time.Second),
	)

	// 文本搜索
	results, err := ddgs.Videos(
		"flower",
		"wt-wt",
		ddg_search.SafeSearchModerate,
		ddg_search.TimelimitAll,
		1,
	)
	if err != nil {
		panic(err)
	}

	for _, r := range results {
		fmt.Printf("videos: %v \n", r)
	}
}
