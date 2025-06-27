package test

import (
	"github.com/Patrick7241/ddg_search"
	"testing"
	"time"
)

func TestNewsDDG(t *testing.T) {
	ddgs := ddg_search.NewDDGS(
		ddg_search.WithProxy("127.0.0.1:7890"), // add proxy
		ddg_search.WithTimeout(10*time.Second),
		ddg_search.WithSleepDuration(10*time.Second),
	)

	// 文本搜索
	results, err := ddgs.News(
		"golang",
		"wt-wt",
		ddg_search.SafeSearchModerate,
		ddg_search.TimelimitAll,
		1,
	)
	if err != nil {
		t.Error(err)
	}

	for _, r := range results {
		t.Logf("news: %v \n", r)
	}
}
