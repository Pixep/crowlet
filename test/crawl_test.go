package crawler

import (
	"testing"
	"time"

	"github.com/Pixep/crowlet/pkg/crawler"
)

func TestMergeCrawlStats(t *testing.T) {
	statsA := crawler.CrawlStats{
		Total:          10,
		StatusCodes:    map[int]int{200: 10},
		Average200Time: time.Duration(1) * time.Second,
		Max200Time:     time.Duration(2) * time.Second,
	}

	statsB := crawler.CrawlStats{
		Total:          6,
		StatusCodes:    map[int]int{200: 2, 404: 4},
		Average200Time: time.Duration(7) * time.Second,
		Max200Time:     time.Duration(9) * time.Second,
	}

	stats := crawler.MergeCrawlStats(statsA, statsB)

	if stats.Total != 16 {
		t.Fatal("Invalid total", stats.Total)
		t.Fail()
	}

	if stats.StatusCodes[200] != 12 ||
		stats.StatusCodes[404] != 4 {
		t.Fatal("Invalid status codes count")
		t.Fail()
	}

	if stats.Average200Time != time.Duration(2)*time.Second {
		t.Fatal("Invalid average 200 time:", stats.Average200Time)
		t.Fail()
	}

	if stats.Max200Time != time.Duration(9)*time.Second {
		t.Fatal("Invalid maximum 200 time:", stats.Max200Time)
		t.Fail()
	}
}
