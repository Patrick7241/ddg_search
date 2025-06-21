package test

import (
	"fmt"
	"github.com/Patrick7241/ddg_search"
	"testing"
	"time"
)

func TestDDG(t *testing.T) {
	//创建带代理的客户端
	ddgs := ddg_search.NewDDGS(
		ddg_search.WithProxy("127.0.0.1:7890"), // add proxy
		ddg_search.WithTimeout(10*time.Second),
		ddg_search.WithSleepDuration(10*time.Second),
	)

	// 文本搜索
	results, err := ddgs.Text(
		"golang",
		"wt-wt",
		ddg_search.SafeSearchModerate,
		ddg_search.TimelimitAll,
		ddg_search.BackendAuto,
		2,
	)
	if err != nil {
		panic(err)
	}

	i := 1
	for _, r := range results {
		fmt.Printf("Title: %s number:%d \n", r["title"], i)
		i++
	}
}
