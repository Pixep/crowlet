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

func AsyncHttpGets(urls []string, user string, pass string) <-chan *HttpResponse {
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
			if len(user) > 0 {
				req.SetBasicAuth(user, pass)
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

			if log.GetLevel() == log.DebugLevel {
				log.WithFields(log.Fields{
					"resp":    resp.StatusCode,
					"dns":     int(result.DNSLookup / time.Millisecond),
					"tcpconn": int(result.TCPConnection / time.Millisecond),
					"tls":     int(result.TLSHandshake / time.Millisecond),
					"server":  int(result.ServerProcessing / time.Millisecond),
					"content": int(result.ContentTransfer(end) / time.Millisecond),
					"close":   end,
				}).Debug("GET: " + url)
			} else {
				log.WithFields(log.Fields{
					"resp":    resp.StatusCode,
					"server":  int(result.ServerProcessing / time.Millisecond),
					"content": int(result.ContentTransfer(end) / time.Millisecond),
				}).Info("GET: " + url)
			}

			ch <- &HttpResponse{url, resp, result, end, err}
		}(url)
	}
	return ch
}
