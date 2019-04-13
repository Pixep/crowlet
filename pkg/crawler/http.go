package crawler

import (
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tcnksm/go-httpstat"
)

// HTTPResponse holds information from a GET to a specific URL
type HTTPResponse struct {
	URL        string
	Response   *http.Response
	Result     *httpstat.Result
	StatusCode int
	EndTime    time.Time
	Err        error
}

// HTTPConfig hold settings used to get pages via HTTP/S
type HTTPConfig struct {
	User    string
	Pass    string
	Timeout time.Duration
}

// HTTPGetter performs a single HTTP/S  to the url, and return information
// related to the result as an HTTPResponse
type HTTPGetter func(url string, config HTTPConfig) (response *HTTPResponse)

func createRequest(url string) (*http.Request, *httpstat.Result, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}

	// create a httpstat powered context
	result := &httpstat.Result{}
	ctx := httpstat.WithHTTPStat(req.Context(), result)
	req = req.WithContext(ctx)

	return req, result, nil
}

func configureRequest(req *http.Request, config HTTPConfig) {
	if len(config.User) > 0 {
		req.SetBasicAuth(config.User, config.Pass)
	}
}

// HTTPGet issues a GET request to a single URL and returns an HTTPResponse
func HTTPGet(url string, config HTTPConfig) (response *HTTPResponse) {
	response = &HTTPResponse{
		URL: url,
	}

	req, result, err := createRequest(url)
	if err != nil {
		response.Err = err
		return
	}

	configureRequest(req, config)

	client := http.Client{
		Timeout: config.Timeout,
	}

	resp, err := client.Do(req)
	defer func() {
		response.EndTime = time.Now()
		response.Response = resp
		response.Result = result

		if resp == nil {
			response.StatusCode = 0
		} else {
			response.StatusCode = response.Response.StatusCode
		}
		PrintResult(response)
	}()

	if err != nil {
		log.Error(err)
		response.Err = err
		return
	}

	// Explicitly Drain & close the body to allow faster
	// reuse of the transport
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	return
}

// ConcurrentHTTPGetter allows concurrent execution of an HTTPGetter
type ConcurrentHTTPGetter interface {
	ConcurrentHTTPGet(urls []string, config HTTPConfig, maxConcurrent int,
		quit <-chan struct{}) <-chan *HTTPResponse
}

// BaseConcurrentHTTPGetter implements HTTPGetter interface using net/http package
type BaseConcurrentHTTPGetter struct {
	Get HTTPGetter
}

// ConcurrentHTTPGet will GET the urls passed and result the results of the crawling
func (getter *BaseConcurrentHTTPGetter) ConcurrentHTTPGet(urls []string, config HTTPConfig,
	maxConcurrent int, quit <-chan struct{}) <-chan *HTTPResponse {

	resultChan := make(chan *HTTPResponse, len(urls))

	go RunConcurrentGet(getter.Get, urls, config, maxConcurrent, resultChan, quit)

	return resultChan
}

// RunConcurrentGet runs multiple HTTP requests in parallel, and returns the
// result in resultChan
func RunConcurrentGet(httpGet HTTPGetter, urls []string, config HTTPConfig,
	maxConcurrent int, resultChan chan<- *HTTPResponse, quit <-chan struct{}) {

	httpResources := make(chan int, maxConcurrent)
	var wg sync.WaitGroup

	defer func() {
		wg.Wait()
		close(resultChan)
	}()

	for _, url := range urls {
		select {
		case <-quit:
			log.Info("Waiting for workers to finish...")
			return
		case httpResources <- 1:
			wg.Add(1)

			go func(url string) {
				defer func() {
					<-httpResources
					wg.Done()
				}()

				resultChan <- httpGet(url, config)
			}(url)
		}
	}
}

// PrintResult will print information relative to the HTTPResponse
func PrintResult(result *HTTPResponse) {
	total := int(result.Result.Total(result.EndTime).Round(time.Millisecond) / time.Millisecond)

	if log.GetLevel() == log.DebugLevel {
		log.WithFields(log.Fields{
			"status":  result.StatusCode,
			"dns":     int(result.Result.DNSLookup / time.Millisecond),
			"tcpconn": int(result.Result.TCPConnection / time.Millisecond),
			"tls":     int(result.Result.TLSHandshake / time.Millisecond),
			"server":  int(result.Result.ServerProcessing / time.Millisecond),
			"content": int(result.Result.ContentTransfer(result.EndTime) / time.Millisecond),
			"time":    total,
			"close":   result.EndTime,
		}).Debug("url=" + result.URL)
	} else {
		log.WithFields(log.Fields{
			"status":     result.StatusCode,
			"total-time": total,
		}).Info("url=" + result.URL)
	}
}
