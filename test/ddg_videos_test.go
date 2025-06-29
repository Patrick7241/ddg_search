package test

import (
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
		"how to learn golang",
		"wt-wt",
		ddg_search.SafeSearchModerate,
		ddg_search.TimelimitAll,
		ddg_search.ResolutionAll,
		ddg_search.DurationAll,
		ddg_search.LicenseAll,
		1,
	)
	if err != nil {
		t.Error(err)
	}

	for _, r := range results {
		t.Logf("videos: %v \n", r)
	}
}
