package main

import (
    "fmt"
		"net/http"
		"log"
		"bytes"
		"math/rand"
		"encoding/json"
		"net/http/httputil"
		"net/url"
		"strings"
		"time"

		"github.com/bradleypeabody/gouuidv6"

		"github.com/Shopify/sarama"
		"github.com/kelseyhightower/envconfig"
		"knative.dev/eventing-contrib/kafka"
		"github.com/go-redis/redis/v8"

	)

type EnvInfo struct {
	Topics   string `envconfig:"KAFKA_TOPICS" required:"true"`
	BootstrapServers []string `envconfig:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
	Key     string `envconfig:"KAFKA_KEY" required:"true"`
	Headers map[string]string `envconfig:"KAFKA_HEADERS" required:"true"`
	Value   string
	RedisMaster string `envconfig:"REDIS_MASTER_NAME" required:"true"`
	Broker string `envconfig:"BROKER" required:"true"`
}

type RequestData struct {
	Method     string  //`json:"method"`
	URL string  //`json:"url"`
	Body   string //`json:"body"`
	ContentType string //`json:"content-type"`
	ID string
}

func main() {
	// Start an HTTP Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// check for Prefer: respond-async header

		var isAsync bool
		isKafka := false

		target := &url.URL{
			Scheme: "http",
			Host: r.Host,
			Path: r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}
		asyncHeader := r.Header.Get("Prefer")
		if asyncHeader == "respond-async" {
			isAsync = true
		}
		if !isAsync {
			proxy := httputil.NewSingleHostReverseProxy(target)
			r.Host = target.Host
			proxy.ServeHTTP(w, r)
		} else {
			// myhost is the url of the app we ultimately want to access
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)
			bodyStr := buf.String()
			id := gouuidv6.NewFromTime(time.Now()).String()
			reqData := RequestData {
				Method: r.Method,
				URL: target.String(),
				Body: bodyStr,
				ID: id,
			}
			reqJSON, err := json.Marshal(reqData)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, "Failed to marshal response: ", err)
				return
			}

			// get env info for kafka
			var s EnvInfo
			err = envconfig.Process("", &s) // BMV TODO: how can we process just a subset of env, providing "kafka" maybe?
			if err != nil {
				log.Fatal(err.Error())
			}
			ctx := r.Context()

			if (isKafka == true) {

				// Create a Kafka client from our Binding.
				client, err := kafka.NewProducer(ctx)
				if err != nil {
					log.Fatal(err.Error())
				}
				producer, err := sarama.NewSyncProducerFromClient(client)
				if err != nil {
					log.Fatal(err.Error())
				}

				// Send the message this Job was created to send.
				headers := make([]sarama.RecordHeader, 0, len(s.Headers))
				for k, v := range s.Headers {
					headers = append(headers, sarama.RecordHeader{
						Key:   []byte(k),
						Value: []byte(v),
					})
				}
				partitionnum := rand.Int31n(20) // NOTE: this does not seem to be effecting the actual partition that gets written
				topicNum := rand.Int31n(10)
				topics := strings.Split(s.Topics, ",")
				partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
					//Topic:   s.Topic,
					Topic:     topics[topicNum],
					// Key:     sarama.StringEncoder(s.Key), //BMV TODO: What does key do?
					Partition: partitionnum,
					Value:   sarama.StringEncoder(reqJSON),
					Headers: headers,
				})
				if err != nil {
					log.Fatal(err.Error())
				} else {
					log.Print(partition)
					log.Print(offset)
					w.WriteHeader(http.StatusAccepted)
				}
			} else {
				opts := &redis.UniversalOptions{
					MasterName: s.RedisMaster,//os.Getenv("REDIS_MASTER_NAME"),
					Addrs:      []string{s.Broker},//[]string{os.Getenv("BROKER")},
				}
				redis := redis.NewUniversalClient(opts)
				fmt.Println("PUSHING ONTO QUEUE")
				rpush := redis.RPush(ctx, "queuename", reqJSON)
				if rpush.Err() != nil  {
					log.Printf("Failed to publish %q %v", reqData.ID, err)
					w.WriteHeader(500)
					fmt.Fprint(w, "Failed to publish task", err)
					return
				}
				w.WriteHeader(http.StatusAccepted)
			}

			// BMV TODO: do we need to close any connections or does writing the header handle this?

		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}