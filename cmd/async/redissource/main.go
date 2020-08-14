/*
Copyright 2018 The Knative Authors

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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"github.com/go-redis/redis/v7"

	// duckv1 "knative.dev/pkg/apis/duck/v1"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
)

type Request struct {
	Method     string  `json:"method"`
	URL string  `json:"url"`
	Body   string `json:"body"`
	ContentType string `json:"content-type"`
}
var (
	eventSource string
	eventType   string
	sink        string
	label       string
	periodStr   string
)

const key = "queuename"

func init() {
	//TODO: update & remove these stringvar?
	flag.StringVar(&eventSource, "eventSource", "", "the event-source (CloudEvents)")
	flag.StringVar(&eventType, "eventType", "dev.knative.eventing.samples.request", "the event-type (CloudEvents)")
	flag.StringVar(&sink, "sink", "", "the host url to write to")
	flag.StringVar(&label, "label", "", "a special label")
	flag.StringVar(&periodStr, "period", "5", "the number of seconds between heartbeats")
}

// BMV TODO: clean up any unrequired config
type envConfig struct {
	// Sink URL where to send heartbeat cloudevents
	Sink string `envconfig:"K_SINK"`

	// CEOverrides are the CloudEvents overrides to be applied to the outbound event.
	CEOverrides string `envconfig:"K_CE_OVERRIDES"`

	// Name of this pod.
	Name string `envconfig:"POD_NAME" required:"true"`

	// Namespace this pod exists in.
	Namespace string `envconfig:"POD_NAMESPACE" required:"true"`

	// Whether to run continuously or exit.
	OneShot bool `envconfig:"ONE_SHOT" default:"false"`
}

func main() {
	opts := &redis.UniversalOptions{
		MasterName: "mymaster",
		Addrs:      []string{"rfs-queue-store:26379"},
	}
	rclient := redis.NewUniversalClient(opts)


   fmt.Println("Waiting for jobs on jobQueue: ", key)
	flag.Parse()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	fmt.Println("ENV SINK", env.Sink)
	if env.Sink != "" {
		sink = env.Sink
	}

	p, err := cloudevents.NewHTTP(cloudevents.WithTarget(sink))
	if err != nil {
		log.Fatalf("failed to create http protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithUUIDs(), cloudevents.WithTimeNow())
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	// var period time.Duration
	// if p, err := strconv.Atoi(periodStr); err != nil {
	// 	period = time.Duration(5) * time.Second
	// } else {
	// 	period = time.Duration(p) * time.Second
	// }

	if eventSource == "" {
		eventSource = fmt.Sprintf("https://knative.dev/eventing-contrib/cmd/heartbeats/#%s/%s", env.Namespace, env.Name)
		log.Printf("Heartbeats Source: %s", eventSource)
	}

	if len(label) > 0 && label[0] == '"' {
		label, _ = strconv.Unquote(label)
	}

	for {

		fmt.Println(time.Now())
		stringdata, err := rclient.BLPop(0*time.Second, key).Result()
		var data Request
		b := []byte(stringdata[1])
		err = json.Unmarshal(b, &data)

		if err != nil {
			log.Fatal(err)
		}

		event := cloudevents.NewEvent("1.0")
		event.SetType(eventType)
		event.SetSource(eventSource)

    go func() {
			if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
				log.Printf("failed to set cloudevents data: %s", err.Error())
			}
	
			log.Printf("sending cloudevent to %s", sink)
			if res := c.Send(context.Background(), event); !cloudevents.IsACK(res) {
				log.Printf("failed to send cloudevent: %v", res)
				fmt.Println("RES", res)
			}
	
			if env.OneShot {
				return
			}
			fmt.Println("END FUNC")
		}()
		fmt.Println("END FOR")
	}
}