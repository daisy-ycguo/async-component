package main

import (
	"testing"
	"flag"
	"bytes"
	"net/http"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"fmt"
)

var (
	eventSource string
	eventType   string
	data        Request
)

func TestConsumeEvent(t *testing.T) {
	t.Run("consume cloud event", func(t *testing.T) {
		myEvent := cloudevents.NewEvent("1.0")
		flag.StringVar(&eventSource, "eventSource", "redis-source", "the event-source (CloudEvents)")
		flag.StringVar(&eventType, "eventType", "dev.knative.async.request", "the event-type (CloudEvents)")
		myEvent.SetType(eventType)
		myEvent.SetSource(eventSource)
		myEvent.SetID("123")

		// BMV TODO: how can we test without using a real URL here?
		req, _ := http.NewRequest("GET", "https://www.google.com/", nil)
		// write the request into b
		var b = &bytes.Buffer{}
		if err := req.Write(b); err !=nil {
			fmt.Println("ERROR WRITING REQUEST")
			// return err
		}
		// translate to string then json with id.
		reqString := b.String() 

		// create data for Request
		data.ID = "123"
		data.Req = reqString
		myEvent.SetData(cloudevents.ApplicationJSON, data)

		theResponse := consumeEvent(myEvent)
		got := theResponse.StatusCode
		want := 200

		if got != want {
				t.Errorf("got %d, want %d", got, want)
		}
})
}