/*
Copyright 2019 The Knative Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Request struct {
	Method     string  `json:"method"`
	URL string  `json:"url"`
	Body   string `json:"body"`
	ContentType string `json:"content-type"`
}

func display(event cloudevents.Event) {
	fmt.Printf("☁️  cloudevents.Event\n%s", event.String())
	data := &Request{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}

	switch data.Method {
	case http.MethodGet:
		resp, err := http.Get(data.URL)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))

	case http.MethodPost:
		resp, err := http.Post(data.URL, data.ContentType, strings.NewReader(data.Body))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))
	}
}

func main() {
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), display))
}