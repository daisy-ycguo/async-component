package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
)

func TestAsyncRequestHeader(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "POST" {
			t.Errorf("Expected 'POST' OR 'GET' request, got '%s'", r.Method)
		}
	}))

	tests := []struct {
		name       string
		async      bool
		method     string
		largeBody  bool
		contentLengthSet bool
		returncode int
	}{{
		name:       "async get request",
		async:      true,
		method:     "GET",
		largeBody:  false,
		contentLengthSet: false,
		returncode: 500, //TODO: how can we test 202 return without standing up redis?
	}, {
		name:       "non async get request",
		async:      false,
		method:     "GET",
		largeBody:  false,
		contentLengthSet: false,
		returncode: 200,
	}, {
		name:       "async post request with too large payload",
		async:      true,
		method:     "POST",
		largeBody:  true,
		contentLengthSet: true,
		returncode: 500,
	}, {
		name:       "async post request with no content-length set",
		async:      true,
		method:     "POST",
		largeBody:  false,
		contentLengthSet: false,
		returncode: 411,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, testserver.URL, nil)
			if test.method == "POST" {
				body := strings.NewReader(`{"body":"this is a body"}`)
				request, _ = http.NewRequest(http.MethodPost, testserver.URL, body)
				if (test.contentLengthSet) {
					if test.largeBody {
						request.Header.Set("Content-Length", "70000000")
					} else {
						request.Header.Set("Content-Length", "1000")
					}
				}
			} 
			if test.async {
				request.Header.Set("Prefer", "respond-async")
			}
			
			rr := httptest.NewRecorder()
			checkHeaderAndServe(rr, request)

			got := rr.Code
			want := test.returncode

			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}
}
