package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
)

func TestAsyncRequestHeader(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "POST" {
			t.Errorf("Expected 'POST' OR 'GET' request, got '%s'", r.Method)
		}
	}))

	tests := []struct {
		name        string
		async   bool
		returncode  int
	}{{
		name:        "make async request",
		async:        true,
		returncode:   200,
	}, {
		name:        "non async request",
		async:       false,
		returncode:  200,
	}}
	// os.Setenv("SOURCE", "redis") //TODO: how to test redis source
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, testserver.URL, nil)
			if (test.async) {
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

	// t.Run("make async request", func(t *testing.T) {
	// 		request, _ := http.NewRequest(http.MethodGet, "/", nil)
	// 		request.Header.Set("Prefer", "respond-async")
	// 		rr := httptest.NewRecorder()

	// 		checkHeaderAndServe(rr, request)

	// 		got := rr.Code
	// 		want := 200

	// 		if got != want {
	// 				t.Errorf("got %d, want %d", got, want)
	// 		}
	// })
}