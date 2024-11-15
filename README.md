# Crowlet

![goreportcard](https://goreportcard.com/badge/github.com/Pixep/crowlet)

Crowlet is a `sitemap.xml` crawler, that can be used as cache warmer, or as a monitoring tool. When ran, it will report information relative to the crawling. Uses include:

- Manual website health check
- Website status and response time monitoring
- Periodic warm up of a server's cache

## Getting Started

The simplest option is to run the tool from its docker image `aleravat/crowlet`. It can otherwise be used from command line as `crowlet`.

### Basic usage

Run the sitemap crawler from its docker image directly

```
docker run -it --rm aleravat/crowlet https://foo.bar/sitemap.xml
```

Or install locally using Go 1.17 or higher:

```
GOBIN=/usr/local/bin go install github.com/Pixep/crowlet/cmd/crowlet@latest
/usr/local/bin/crowlet https://foo.bar/sitemap.xml
```

Or build locally with `make build` and run.
You can alternatively use `make build-static-linux` to build a self contained binary for Linux platforms.

```
make build
./crowlet https://foo.bar/sitemap.xml
```

### Use scenarios

Crowlet can be used in a few different ways, as described below.

#### Check website's status

Ran once on a sitemap, it will provide information regarding return codes and complete server time.

```
crowlet https://google.com/sitemap.xml
INFO[0000] Crawling https://google.com/sitemap.xml
INFO[0020] Found 5010 URL(s)
INFO[0020] URL: https://www.google.com/intl/ar/gmail/about/for-work/  status=200 total-time=85
INFO[0020] URL: https://www.google.com/intl/ar/gmail/about/  status=200 total-time=86
INFO[0020] URL: https://www.google.com/intl/am/gmail/about/for-work/  status=200 total-time=87
INFO[0020] URL: https://www.google.com/intl/am/gmail/about/policy/  status=200 total-time=87
INFO[0020] URL: https://www.google.com/intl/am/gmail/about/  status=200 total-time=88
[...]
INFO[0021] -------- Summary -------
INFO[0021] general:
INFO[0021]     crawled: 51
INFO[0021]
INFO[0021] status:
INFO[0021]     status-200: 51
INFO[0021]
INFO[0021] server-time:
INFO[0021]     avg-time: 61ms
INFO[0021]     max-time: 145ms
INFO[0021] ------------------------
```

#### Cache warmer

You can use this tool as to warm cache for all URLs in a sitemap using the `--forever` option. This will keep crawling the sitemap forever, and `--wait-interval` can be used to define the pause duration in seconds, between each complete crawling.

```bash
# Crawl the sitemap every 30 minutes
$ docker run -it --rm aleravat/crowlet --forever --wait-interval 1800 https://foo.bar/sitemap.xml
```

#### Status monitoring

If any page from the sitemap returns a non `200` status code, crowlet will return with exit code `1`. This can be used and customized to monitor the status of the pages, and automate error detection. The `--non-200-error` option allow setting the exit code if any page has a non `200` status code.

```bash
# Return with code `150` if any page has a status != 200
docker run -it --rm aleravat/crowlet --non-200-error 150 https://foo.bar/sitemap.xml
```

The `--json` flag can be used, as well as `--summary-only` for an easy parsing of the output.

```
./crowlet --json --summary-only https://google.com/sitemap.xml
{"total":{"crawled":43},"status":{"status-codes":{"200":43},"errors":null},"response-time":{"avg-time-ms":87,"max-time-ms":418}}
```

The `--crawl-images`, `--crawl-hyperlinks` and `--crawl-external` options can be used to extends the monitoring to internal (or even external) links found in the original sitemap pages. Their statistics will be added to the final report.

#### Response time monitoring

The `--response-time-max` option can be used to indicate a maximum server total time, or crowlet will return with `--response-time-error` return code. Note that if any page return a status code different from 200, the `--non-200-error` code will be returned instead.

```bash
# Return with code `5` if any page takes more than `1000`ms until reception
# -t 1: Load pages one by one to avoid biased measurement
docker run -it --rm aleravat/crowlet -t 1 -l 5 -m 1000 https://foo.bar/sitemap.xml
```

### Command line options

The following arguments can be used to customize it to your needs:

```
COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --crawl-hyperlinks                     follow and test hyperlinks ('a' tags href)
   --crawl-images                         follow and test image links ('img' tags src)
   --crawl-external                       follow and test external links. Use in combination with 'follow-hyperlinks' and/or 'follow-images'
   --forever, -f                          crawl the sitemap's URLs forever... or until stopped
   --iterations value, -i value           number of crawling iterations for the whole sitemap (default: 1)
   --wait-interval value, -w value        wait interval in seconds between sitemap crawling iterations (default: 0) [$CRAWL_WAIT_INTERVAL]
   --throttle value, -t value             number of http requests to do at once (default: 5) [$CRAWL_THROTTLE]
   --timeout value, -y value              timeout duration for requests, in milliseconds (default: 20000)
   --quiet, --silent, -q                  suppress all normal output
   --json, -j                             output using JSON format (experimental)
   --non-200-error value, -e value        error code to use if any non-200 response if encountered (default: 1)
   --response-time-error value, -l value  error code to use if the maximum response time is overrun (default: 1)
   --response-time-max value, -m value    maximum response time of URLs, in milliseconds, before considered an error (default: 0)
   --summary-only                         print only the summary
   --override-host value                  override the hostname used in sitemap urls [$CRAWL_HOST]
   --user value, -u value                 username for http basic authentication [$CRAWL_HTTP_USER]
   --pass value, -p value                 password for http basic authentication [$CRAWL_HTTP_PASSWORD]
   --pre-cmd value                        command(s) to run before starting crawler
   --post-cmd value                       command(s) to run after crawler finishes
   --debug                                run in debug mode
   --help, -h                             show help
   --version, -v                          print the version
```

## License

This project is licensed under the Apache-2.0 License- see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgment

Thanks to @chris-scentregroup for his great initial work with `sitemap-crawler`
