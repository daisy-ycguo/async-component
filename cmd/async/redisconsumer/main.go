package main

import (
   "fmt"
   "github.com/go-redis/redis/v7"
   "log"
	 "time"
	 "encoding/json"
	 "io/ioutil"
	 "net/http"
	 "strings"

	//  "os"
)

type Request struct {
	Method     string  `json:"method"`
	URL string  `json:"url"`
	Body   string `json:"body"`
	ContentType string `json:"content-type"`
}

const key = "queuename"

func main() {

  //  c := redis.NewClient(&redis.Options{
  //     Addr: "localhost:6379",
	//  })
	 opts := &redis.UniversalOptions{
		MasterName: "mymaster",
		Addrs:      []string{"rfs-queue-store:26379"},
	}
	c := redis.NewUniversalClient(opts)


   fmt.Println("Waiting for jobs on jobQueue: ", key)

   go func() {
      for {
				fmt.Println("IN FOR")
				 stringdata, err := c.BLPop(0*time.Second, key).Result()
				 var data Request
				 b := []byte(stringdata[1])
				 err = json.Unmarshal(b, &data)

				 fmt.Println("RESULT?", data)

         if err != nil {
            log.Fatal(err)
         }

				 fmt.Println("Executing job: ", data)
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
   }()

   // block for ever, used for testing only
   select {}
}