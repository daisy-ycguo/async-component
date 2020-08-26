package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"context"
	"time"

	"github.com/bradleypeabody/gouuidv6"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"knative.dev/eventing-contrib/kafka"
)

type EnvInfo struct {
	Topics           string            `envconfig:"KAFKA_TOPICS"`
	BootstrapServers []string          `envconfig:"KAFKA_BOOTSTRAP_SERVERS"`
	Key              string            `envconfig:"KAFKA_KEY"`
	Headers          map[string]string `envconfig:"KAFKA_HEADERS"`
	Source           string            `envconfig:"SOURCE"`
	RedisMaster      string            `envconfig:"REDIS_MASTER_NAME"`
	Broker           string            `envconfig:"BROKER"`
	Value            string
}

type RequestData struct {
	ID          string //`json:"id"`
	Request     string //`json:"request"`
}

func main() {
	// Start an HTTP Server
	http.HandleFunc("/", checkHeaderAndServe)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkHeaderAndServe(w http.ResponseWriter, r *http.Request) {
	// check for Prefer: respond-async header
	var isAsync bool

	target := &url.URL{
		Scheme:   "http",
		Host:     r.Host,
		Path:     r.URL.Path,
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
		// write the request into b
		var b = &bytes.Buffer{}
		if err := r.Write(b); err !=nil {
			fmt.Println("ERROR WRITING REQUEST")
			// return err
		}
		// translate to string then json with id.
		reqString := b.String() 
		id := gouuidv6.NewFromTime(time.Now()).String()
		reqData := RequestData{
			ID: id,
			Request:    reqString,
		}
		reqJSON, err := json.Marshal(reqData)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Failed to marshal request: ", err)
			return
		}

		// get env info for queue
		var s EnvInfo
		err = envconfig.Process("", &s) // BMV TODO: how can we process just a subset of env, providing "kafka" maybe?
		if err != nil {
			log.Fatal(err.Error())
		}
		ctx := r.Context()
		if s.Source == "kafka" {
			writeToKafka(ctx, s, reqJSON, w)
		} else if s.Source == "redis" {
			writeToRedis(ctx, s, reqJSON, w, reqData.ID)
		}

		// BMV TODO: do we need to close any connections or does writing the header handle this?

	}
}

func writeToKafka(ctx context.Context, s EnvInfo, reqJSON []byte, w http.ResponseWriter) {
	fmt.Println("USING KAFKA")
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
	partitionnum := rand.Int31n(3)
	topicNum := rand.Int31n(10)
	topics := strings.Split(s.Topics, ",")
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		//Topic:   s.Topic,
		Topic: topics[topicNum],
		// Key:     sarama.StringEncoder(s.Key), //BMV TODO: What does key do?
		Partition: partitionnum,
		Value:     sarama.StringEncoder(reqJSON),
		Headers:   headers,
	})
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Print(partition)
		log.Print(offset)
		w.WriteHeader(http.StatusAccepted)
	}
}

func writeToRedis(ctx context.Context, s EnvInfo, reqJSON []byte, w http.ResponseWriter, id string) {
	fmt.Println("USING REDIS")
	opts := &redis.UniversalOptions{
		Addrs: []string{"redis.redis.svc.cluster.local:6379"},
	}
	client := redis.NewUniversalClient(opts)
	fmt.Println("PUSHING ONTO QUEUE", reqJSON)
	// TODO: maybe there's a different way to add this to stream?
	strCMD := client.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{
				"data": reqJSON,
		},
	})
	// rpush := client.RPush(ctx, "queuename", reqJSON)
	if strCMD.Err() != nil {
		log.Printf("Failed to publish %q %v", id, strCMD.Err())
		w.WriteHeader(500)
		fmt.Fprint(w, "Failed to publish task", strCMD.Err())
		return
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}
