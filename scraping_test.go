// Package scraping gives a middleware that scraps data from given URLs
package scraping

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	mockSrv := httptest.NewServer(http.HandlerFunc(restHandler(t)))
	defer mockSrv.Close()

	http.Handle("/", Handler())

	tests := []struct {
		name           string
		path           string
		method         string
		contentType    string
		body           io.Reader
		wantStatusCode int
	}{
		{
			name:           "correct response for 2 sites",
			path:           "/",
			contentType:    "text/plain",
			method:         http.MethodPost,
			body:           strings.NewReader(fmt.Sprintf("%s/google\n%s/example", mockSrv.URL, mockSrv.URL)),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "not found response for incorrect path",
			path:           "/not-found",
			contentType:    "text/plain",
			method:         http.MethodPost,
			body:           nil,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "unsupported media response with illegal content type",
			path:           "/",
			contentType:    "application/json",
			method:         http.MethodPost,
			body:           nil,
			wantStatusCode: http.StatusUnsupportedMediaType,
		},
		{
			name:           "method not allowed with illegal method",
			path:           "/",
			contentType:    "text/plain",
			method:         http.MethodGet,
			body:           nil,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()
			handler := Handler()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("got status %d but wanted %d", rr.Code, tt.wantStatusCode)
			}
			// Check the response body is what we expect.
			// expected := `{"alive": true}`
			// if rr.Body.String() != expected {
			// 	t.Errorf("handler returned unexpected body: got %v want %v",
			// 		rr.Body.String(), expected)
			// }
		})
	}
}

func Test_unsupportedContentType(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "return correct status code",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsupportedContentType(tt.args.w, tt.args.r)
			if tt.args.w.Code != http.StatusUnsupportedMediaType {
				t.Errorf("got status %d but wanted %d", tt.args.w.Code, http.StatusUnsupportedMediaType)
			}
		})
	}
}

func Test_notAllowed(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "return correct response",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notAllowed(tt.args.w, tt.args.r)
			if tt.args.w.Code != http.StatusMethodNotAllowed {
				t.Errorf("got status %d but wanted %d", tt.args.w.Code, http.StatusMethodNotAllowed)
			}
			allow := tt.args.w.Header().Get("Allow")
			if allow != "POST, OPTIONS" {
				t.Errorf("got header %s but wanted %s", allow, "POST, OPTIONS")
			}
		})
	}
}

func Test_getData(t *testing.T) {
	type args struct {
		url   string
		resCh chan<- int
		errCh chan<- error
	}
	resCh := make(chan int)
	errCh := make(chan error)
	mockSrv := httptest.NewServer(http.HandlerFunc(restHandler(t)))

	defer func() {
		mockSrv.Close()
		close(resCh)
		close(errCh)
	}()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "get data from the site",
			args: args{url: mockSrv.URL + "/google", resCh: resCh, errCh: errCh},
		},
		{
			name:    "get err from an invalid url",
			args:    args{url: " ", resCh: resCh, errCh: errCh},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go getData(tt.args.url, tt.args.resCh, tt.args.errCh)

			select {
			case data := <-resCh:
				if data == 0 {
					t.Error("expected data")
				}
			case err := <-errCh:
				if !tt.wantErr {
					t.Error("error was not expected", err)
				}
			case <-time.After(5 * time.Second):
				t.Error("timeout")
			}
		})
	}
}

func Test_processURLs(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	mockSrv := httptest.NewServer(http.HandlerFunc(restHandler(t)))
	defer mockSrv.Close()

	twoSites := strings.NewReader(fmt.Sprintf("%s/google\n%s/example", mockSrv.URL, mockSrv.URL))
	invalidSite := strings.NewReader(" ")

	tests := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "correct response for 2 sites",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/", twoSites),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "correct response for invalid body",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/", invalidSite),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processURLs(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatusCode {
				t.Fatalf("got status %d but wanted %d", tt.args.w.Code, http.StatusOK)
			}
		})
	}
}

func restHandler(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimSpace(r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		data, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.html", p))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(data)
	}
}
