# sitemap-crawler

This tool will crawl a sitemap, and report information relative to the crawling. It can be used with a sitemap to:
- Test URLs for errors
- Warm up server's cache

## Getting Started

Once built with `go install`, you will be able to use `smapcrawl` application from command line.

### Command line options

The following arguments can be used to customize its behavior:
```
   --host value                     override the hostname used in sitemap urls [$CRAWL_HOST]
   --user value, -u value           username for http basic authentication [$CRAWL_HTTP_USER]
   --pass value, -p value           password for http basic authentication [$CRAWL_HTTP_PASSWORD]
   --forever, -f                    reads the sitemap once keep crawling all urls until stopped
   --wait-interval value, -w value  wait interval in seconds between sitemap crawling iterations (default: 0) [$CRAWL_WAIT_INTERVAL]
   --throttle value, -t value       number of http requests to do at once (default: 5) [$CRAWL_THROTTLE]
   --pre-cmd value                  command(s) to run before starting crawler
   --post-cmd value                 command(s) to run after crawler finishes
   --debug, -d                      run in debug mode
   --help, -h                       show help
   --version, -v                    print the version
```

## License

This project is licensed under the Apache-2.0 License- see the [LICENSE.md](LICENSE.md) file for details
