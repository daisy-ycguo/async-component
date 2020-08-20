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
	"strings"
	"bufio"
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Request struct {
	ID      string `json:"id"`
	Req string `json:"request"`
}

func consumeEvent(event cloudevents.Event) error {
	fmt.Printf("☁️  cloudevents.Event\n%s", event.String())

	data := &Request{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
		return err
	}

	fmt.Println("REQUEST DATA", data.Req)
	r := bufio.NewReader(strings.NewReader(data.Req))
	var req *http.Request
	var err error
	if req, err = http.ReadRequest(r); err != nil { // deserialize request
		fmt.Println("Problem reading request: T", err)
		return err
	}
	// client for sending request
	client := &http.Client{}

	// build new url - writing the request removes the URL and places in URI.
	req.RequestURI = ""
	req.URL, _ = url.Parse("http://" + req.Host + req.RequestURI) //TODO: catch this error later
	req.Header.Del("Prefer") // We do not want to make this request as async
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Problem calling url: ", err)
		return err
	}
	defer resp.Body.Close()
	// read body from response
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(body)
	return nil
}

func main() {
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), consumeEvent))
}
