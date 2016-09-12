package nrinsights_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"dcx.rax.io/nrinsights"
)

//Event is a implementation of NREvent
type Event struct {
	EventType string `json:"eventType"`
	EventName string `json:"eventName"`
	Timestamp int64  `json:"timestamp"`
}

func ExampleClient() {
	ch := make(chan interface{}, 1)

	// Construct listener server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		//fmt.Println("Success!")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ch <- err
		}
		ch <- b
	}))
	defer ts.Close()

	// Create a new client with NewRelic token + collection endpoint
	cfg := nrinsights.DefaultConfig
	cfg.Endpoint = ts.URL
	cfg.Token = "fake_token"
	c, err := nrinsights.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Generate fake data
	message := Event{"EventCreation", "Foo", 1461097120}

	// Send data to listener
	c.Send(message)

	// Close to force a batch out
	c.Close()

	// Wait to ensure message is sent
	fmt.Printf("%s", <-ch)

	// Output: [{"eventType":"BlockCreation","eventName":"Foo","timestamp":1461097120}]
}
