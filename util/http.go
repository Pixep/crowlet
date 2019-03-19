package util

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tcnksm/go-httpstat"
)

type HttpResponse struct {
	URL      string
	Response *http.Response
	Result   httpstat.Result
	EndTime  time.Time
	Err      error
}

// HTTPConfig hold settings used to get pages via HTTP/S
type HTTPConfig struct {
	User string
	Pass string
}

func AsyncHttpGets(urls []string, config HTTPConfig) <-chan *HttpResponse {
	ch := make(chan *HttpResponse, len(urls)) // buffered
	for _, url := range urls {
		go func(url string) {

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Fatal(err)
			}

			// create a httpstat powered context
			var result httpstat.Result
			ctx := httpstat.WithHTTPStat(req.Context(), &result)
			req = req.WithContext(ctx)

			// set http basic if provided
			if len(config.User) > 0 {
				req.SetBasicAuth(config.User, config.Pass)
			}

			client := http.DefaultClient
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			// Explicitly Drain & close the body to allow faster
			// reuse of the transport
			defer func() {
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
			}()

			end := time.Now()
			total := int(result.Total(end).Round(time.Millisecond) / time.Millisecond)

			if log.GetLevel() == log.DebugLevel {
				log.WithFields(log.Fields{
					"status":  resp.StatusCode,
					"dns":     int(result.DNSLookup / time.Millisecond),
					"tcpconn": int(result.TCPConnection / time.Millisecond),
					"tls":     int(result.TLSHandshake / time.Millisecond),
					"server":  int(result.ServerProcessing / time.Millisecond),
					"content": int(result.ContentTransfer(end) / time.Millisecond),
					"time":    total,
					"close":   end,
				}).Debug("url=" + url)
			} else {
				log.WithFields(log.Fields{
					"status":     resp.StatusCode,
					"total-time": total,
				}).Info("url=" + url)
			}

			ch <- &HttpResponse{url, resp, result, end, err}
		}(url)
	}
	return ch
}
