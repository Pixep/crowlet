package util

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tcnksm/go-httpstat"
	log "github.com/Sirupsen/logrus"
)

type HttpResponse struct {
	Url      string
	Response *http.Response
	Err      error
}

func AsyncHttpGets(urls []string, user string, pass string) <-chan *HttpResponse {
	ch := make(chan *HttpResponse, len(urls)) // buffered
	for _, url := range urls {
		go func(url string) {
			//log.Info("Fetching " + url)

			// create a new http request
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

			// send request by default http client
			//client := &http.Client{}
			// resp, err := client.Do(req)
			// if err == nil {
			// 	resp.Body.Close()
			// }

			client := http.DefaultClient
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
				log.Fatal(err)
			}
			if err == nil {
				resp.Body.Close()
			}
			end := time.Now()

			log.WithFields(log.Fields{
				"resp":    resp.StatusCode,
				"server":  int(result.ServerProcessing / time.Millisecond),
				"content": int(result.ContentTransfer(time.Now()) / time.Millisecond),
			}).Info("GET: " + url)

			log.WithFields(log.Fields{
				"resp":    resp.StatusCode,
				"dns":     int(result.DNSLookup / time.Millisecond),
				"tcpconn": int(result.TCPConnection / time.Millisecond),
				"tls":     int(result.TLSHandshake / time.Millisecond),
				"server":  int(result.ServerProcessing / time.Millisecond),
				"content": int(result.ContentTransfer(time.Now()) / time.Millisecond),
				"close":   end,
			}).Debug("GET: " + url)

			ch <- &HttpResponse{url, resp, err}
		}(url)
	}
	return ch
}
