package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/urfave/cli"
)

func BuildSitemap(host string, count int) string {
	sitemap := `<?xml version="1.0" encoding="UTF-8"?>{toto}
	<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	for i := 1; i <= count; i++ {
		if i%10 == 0 {
			sitemap += fmt.Sprintf("<url><loc>%s/error%d</loc></url>", host, i)
		} else {
			sitemap += fmt.Sprintf("<url><loc>%s/page%d</loc></url>", host, i)
		}
	}
	sitemap += "</urlset>"
	return sitemap
}

func BuildPageContent(size int) string {
	content := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Test Page</title>
</head>
<body>
    <h1>Test Page</h1>
    <p>This is a test page with repeated content to reach approximately 2KB in size.</p>
`
	paragraph := `<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus lacinia odio vitae vestibulum vestibulum. Cras venenatis euismod malesuada.</p>`

	// Append paragraphs until the content reaches the desired size
	for len(content) < size {
		content += paragraph
	}
	content += `
</body>
</html>`
	return content
}

func BenchmarkStartFunction(b *testing.B) {
	pageContent := BuildPageContent(150000)
	var sitemapXML string

	// Set up a mock HTTP server to simulate sitemap and page responses
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sitemap.xml" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sitemapXML))
		} else if strings.HasPrefix(r.URL.Path, "/page") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(pageContent))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	sitemapXML = BuildSitemap(mockServer.URL, 1000)

	defer mockServer.Close()

	// Step 2: Set up CLI flags and arguments for the context
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	set.Int("throttle", 3, "number of http requests to do at once")
	set.String("override-host", "", "override hostname")
	set.String("user", "", "username for basic auth")
	set.String("pass", "", "password for basic auth")
	set.Int("timeout", 1000, "timeout in milliseconds")
	set.Bool("crawl-external", false, "crawl external links")
	set.Bool("crawl-images", false, "crawl images")
	set.Bool("crawl-hyperlinks", true, "crawl hyperlinks")
	set.Int("iterations", 1, "number of crawl iterations")
	set.Bool("forever", false, "crawl forever")
	set.Int("wait-interval", 0, "wait interval between iterations")
	set.Bool("quiet", true, "suppress output")
	set.Bool("json", false, "json output")
	set.Int("non-200-error", 1, "error code for non-200 responses")
	set.Int("response-time-error", 2, "error code for max response time exceeded")
	set.Int("response-time-max", 0, "max response time in milliseconds")
	set.Bool("summary-only", false, "only print summary")

	// Add sitemap URL as the argument
	set.Parse([]string{mockServer.URL + "/sitemap.xml"})

	// Create context with flags and args
	ctx := cli.NewContext(app, set, nil)

	// Start the benchmark test
	b.ResetTimer()   // Reset the timer to measure only the time spent in the loop
	b.ReportAllocs() // Report memory allocations per operation

	for i := 0; i < b.N; i++ {
		err := start(ctx)
		if err != nil {
			b.Fatalf("start function failed: %v", err)
		}
	}
}
