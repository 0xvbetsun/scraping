// Package scraping gives a middleware that scraps data from given URLs
package scraping

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

// MaxConnections it's maximum of simultaneous connections
const MaxConnections uint32 = 999

// Handler implements a simple handler for scrapping data from given urls
func Handler() http.Handler {
	var connects uint32

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)

			return
		}
		if r.Method != http.MethodPost {
			notAllowed(w, r)

			return
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			unsupportedContentType(w, r)

			return
		}
		if connects == MaxConnections {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)

			return
		}
		atomic.AddUint32(&connects, 1)
		defer atomic.AddUint32(&connects, ^uint32(0))

		processURLs(w, r)
	})
}

func notAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "POST, OPTIONS")
	msg := fmt.Sprintf("Method %s is not allowed.", r.Method)
	http.Error(w, msg, http.StatusMethodNotAllowed)
}

func unsupportedContentType(w http.ResponseWriter, r *http.Request) {
	msg := "Content-Type header is not text/plain"
	http.Error(w, msg, http.StatusUnsupportedMediaType)
}

func processURLs(w http.ResponseWriter, r *http.Request) {
	numURLs := 0
	resCh := make(chan int)
	errCh := make(chan error)
	defer func() {
		close(resCh)
		close(errCh)
	}()

	sc := bufio.NewScanner(r.Body)
	for sc.Scan() {
		go getData(sc.Text(), resCh, errCh)
		numURLs++
	}

	var buf bytes.Buffer

	for i := 0; i < numURLs; i++ {
		select {
		case size, more := <-resCh:
			if !more {
				// handle writing into the closed channel
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}
			if i == 0 {
				buf.WriteString(strconv.Itoa(size))
			} else {
				buf.WriteString(LineBreak + strconv.Itoa(size))
			}
		case err, more := <-errCh:
			if !more {
				// handle writing into the closed channel
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}
			msg := fmt.Sprintf("Error was ocurred: %s", err)
			http.Error(w, msg, http.StatusInternalServerError)

			return
		case <-time.After(10 * time.Second):
			// handle timeout
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		http.Error(w, "Failed to send out response", http.StatusInternalServerError)
	}
}

func getData(url string, resCh chan<- int, errCh chan<- error) {
	resp, err := http.Get(url)
	if err != nil {
		errCh <- fmt.Errorf("failed to send request: %w", err)

		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errCh <- fmt.Errorf("failed to read body: %w", err)

		return
	}

	resCh <- len(body)
}
