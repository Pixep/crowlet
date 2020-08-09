package crawler

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type summary struct {
	General          generalInfo      `json:"total"`
	StatusInfo       statusInfo       `json:"status"`
	ResponseTimeInfo responseTimeInfo `json:"response-time"`
}

type generalInfo struct {
	Total int `json:"crawled"`
}

type statusInfo struct {
	StatusCodes map[int]int   `json:"status-codes"`
	Non200Urls  []CrawlResult `json:"errors"`
}

type responseTimeInfo struct {
	AverageTimeMs int `json:"avg-time-ms"`
	MaxTimeMs     int `json:"max-time-ms"`
}

// PrintJSONSummary prints a summary of HTTP response codes in JSON format
func PrintJSONSummary(stats CrawlStats) {
	summary := summary{
		General: generalInfo{
			Total: stats.Total,
		},
		StatusInfo: statusInfo{
			StatusCodes: stats.StatusCodes,
			Non200Urls:  stats.Non200Urls,
		},
		ResponseTimeInfo: responseTimeInfo{
			AverageTimeMs: int(stats.Average200Time / time.Millisecond),
			MaxTimeMs:     int(stats.Max200Time / time.Millisecond),
		}}

	jsonSummary, err := json.Marshal(summary)
	if err != nil {
		log.Error("Error generating JSON summary:", err)
		return
	}

	println(string(jsonSummary))
}

// PrintSummary prints a summary of HTTP response codes
func PrintSummary(stats CrawlStats) {
	log.Info("-------- Summary -------")
	log.Info("general:")
	log.Info("    crawled: ", stats.Total)
	log.Info("")
	log.Info("status:")
	for code, count := range stats.StatusCodes {
		log.Info("    status-", code, ": ", count)
	}

	log.Info("")
	log.Info("status-errors-detail:")
	if len(stats.Non200Urls) == 0 {
		log.Info("    - none")
	} else {
		for _, crawlResult := range stats.Non200Urls {
			log.Info("    - ", crawlResult.URL, ":")
			log.Info("        status-code: ", crawlResult.StatusCode)
			for _, linkingURL := range crawlResult.LinkingURLs {
				log.Info("        linking-url: ", linkingURL)
			}
		}
	}

	log.Info("")
	log.Info("server-time: ")
	log.Info("    avg-time: ", int(stats.Average200Time/time.Millisecond), "ms")
	log.Info("    max-time: ", int(stats.Max200Time/time.Millisecond), "ms")
	log.Info("------------------------")
}
