package nrinsights

import (
	"flag"
	"fmt"
	"time"
)

var ErrNoEndpoint = fmt.Errorf("no endpoint configured")
var ErrNoToken = fmt.Errorf("no token configured")
var ErrDuplicateSetup = fmt.Errorf("global client already setup")
var ErrBatchSize = fmt.Errorf("MaxBatchSize must by 1 or larger")
var ErrBatchDelay = fmt.Errorf("MaxBatchDelay must by 1 second or longer")

// Client exposes methods to post messages to New Relic Insights
type Client interface {
	Send(message NREvent)
	Close()
}

// NREvent is the generic type for an event to be sent
type NREvent interface{}

// Config holds the configuration data for nrinsights
type Config struct {
	Endpoint      string
	Token         string
	MaxBatchSize  int
	MaxBatchDelay time.Duration
}

var globalCfg Config

// DefaultConfig is a convenience structure to get default configuration values.
var DefaultConfig = Config {
	Endpoint:      "",
	Token:         "",
	MaxBatchSize:  100,
	MaxBatchDelay: time.Minute,
}

// RegisterFlags registers configuration flags directly, as an alternative to providing them to the NewClient factory
// If called, use NewClientFromFlags or SetupGlobalClientFromFlags.
func RegisterFlags() {
	flag.StringVar(&globalCfg.Endpoint, "nrinsights-endpoint", DefaultConfig.Endpoint, "New Relic Insights API endpoint")
	flag.StringVar(&globalCfg.Token, "nrinsights-token", DefaultConfig.Token, "New Relic Insights API access token")
	flag.IntVar(&globalCfg.MaxBatchSize, "nrinsights-batch-size", DefaultConfig.MaxBatchSize, "Maximum size of message batches to Insights endpoint")
	flag.DurationVar(&globalCfg.MaxBatchDelay, "nrinsights-batch-delay", DefaultConfig.MaxBatchDelay, "Maximum delay between batches to Insights endpoint")
}
