package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
)

func TestAsyncRequestHeader(t *testing.T) {
	t.Run("make async request", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Prefer", "respond-async")
			rr := httptest.NewRecorder()

			checkHeaderAndServe(rr, request)

			got := rr.Code
			want := 200

			if got != want {
					t.Errorf("got %d, want %d", got, want)
			}
	})
}