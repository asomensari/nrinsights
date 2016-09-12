## Synopsis

NRInsights is a library for sending custom events to NewRelic Insights from a Go application. Documentation for this api can be found [here](https://docs.newrelic.com/docs/insights/new-relic-insights/adding-querying-data/insert-custom-events-insights-api)

## Code Example
To send events, you need to create a client. You can create a global singleton client, or a client for a single calls

```
  type Event struct {
	EventType string `json:"eventType"`
	EventName string `json:"eventName"`
	Timestamp int64  `json:"timestamp"`
}


cfg := nrinsights.DefaultConfig
cfg.Endpoint = ts.URL
cfg.Token = "newrelic_token"
c, err := nrinsights.NewClient(cfg)
if err != nil {
  log.Fatal(err)
}

message := Event{"EventCreation", "Foo", 1461097120}

c.Send(message)
```


## Tests

The single test is in nr_test.go. You can run it with `go test`


## License

This project is released under the unlicense
