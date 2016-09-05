package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

var UserAgent = fmt.Sprintf("mcls/hcheck; %s", runtime.Version())

func main() {
	timeout := flag.Duration("timeout", 1000*time.Millisecond, "request timeout in ms")
	printErrorsOnly := flag.Bool("errors-only", false, "only print errors")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] [urls]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Fprint(os.Stderr, "error: no urls specified\n")
		flag.Usage()
	}
	urls := flag.Args()

	client := &http.Client{
		Timeout: *timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}

			// Prevent default User-Agent from being used
			// See: https://github.com/golang/go/issues/4800
			req.Header.Set("User-Agent", UserAgent)
			return nil
		},
	}

	for result := range healthcheck(client, urls) {
		if *printErrorsOnly && result.Success() {
			continue
		}
		printResult(result)
	}
}

func healthcheck(client *http.Client, urls []string) chan checkResult {
	results := make(chan checkResult, len(urls))
	go func() {
		var wg sync.WaitGroup
		for _, url := range urls {
			wg.Add(1)
			go func(url string, results chan checkResult) {
				defer wg.Done()
				result := checkResult{URL: url}
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					result.Error = err
					results <- result
					return
				}
				req.Header.Set("User-Agent", UserAgent)
				start := time.Now()
				resp, err := client.Do(req)
				result.Response = resp
				result.Error = err
				result.Duration = time.Since(start)
				results <- result
			}(url, results)
		}
		wg.Wait()
		close(results)
	}()

	return results
}

type checkResult struct {
	URL      string
	Response *http.Response
	Error    error
	Duration time.Duration
}

func printResult(result checkResult) {
	err := result.Error
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Printf("timeout: %s \n", result.URL)
		} else {
			fmt.Printf("error: %v (%s) \n", err, result.URL)
		}
		return
	}
	fmt.Printf(
		"%s (%dms) - %s\n",
		result.Response.Status,
		result.Duration/time.Millisecond,
		result.URL,
	)
}

// Success returns true if the response status code is within the 2xx range and
// no error occurred
func (r *checkResult) Success() bool {
	if r.Error != nil {
		return false
	}
	code := r.Response.StatusCode
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}
