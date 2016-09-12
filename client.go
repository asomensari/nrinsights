package nrinsights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// client is struct that holds the newrelic details.  It fulfills the Client interface
type client struct {
	h *handler
}

// handler handles messages and sends them off to new relic in batches
type handler struct {
	cfg    Config
	wg     *sync.WaitGroup
	msgCh  chan NREvent
	quitCh chan struct{}
	doneCh chan struct{}
}

// NewClient creates a new client structure using the provided details, returning it as a Client interface
// If either the endpoint or token are empty, an error will be returned, and the returned Client will silently no-op
// return on all methods.
func NewClient(cfg Config) (Client, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return &nilCheckClient{}, ErrNoEndpoint
	}
	if strings.TrimSpace(cfg.Token) == "" {
		return &nilCheckClient{}, ErrNoToken
	}
	if cfg.MaxBatchSize < 1 {
		return &nilCheckClient{}, ErrBatchSize
	}
	if cfg.MaxBatchDelay < time.Second {
		return &nilCheckClient{}, ErrBatchDelay
	}

	cli := &client{
		h: &handler{
			cfg:    cfg,
			wg:     new(sync.WaitGroup),
			msgCh:  make(chan NREvent),
			quitCh: make(chan struct{}),
			doneCh: make(chan struct{}),
		},
	}
	go cli.h.manager()
	return cli, nil
}

// NewClientFromFlags creates a new client structure using the parsed flags, returning it as a Client interface
// If either the endpoint or token are empty, an error will be returned, and the returned Client will silently no-op
// return on all methods.
func NewClientFromFlags() (Client, error) {
	return NewClient(globalCfg)
}

// Send takes a generic NREvent and adds it to the messages channel
func (c *client) Send(message NREvent) {
	select {
	case <-c.h.quitCh: // if we're shutting down, abort the message send
	default:
		c.h.msgCh <- message
	}
}

// Close terminates the handler.  All existing worker routines will send their batches to new relic, then exit.
// Close will not return until all of the workers have written their batches to new relic and properly shut down.
func (c *client) Close() {
	c.h.printDebug("Close() called.")
	select {
	case <-c.h.quitCh: // makes this idempotent, since closing a closed channel panics
	default:
		close(c.h.quitCh)
	}
	c.h.printDebug("quitCh closed.")
	<-c.h.doneCh
	c.h.printDebug("Close() complete.")
}

// manager handles incoming messages over the message channel, buffering the and sending them to the Insights endpoint
// one the buffer reaches the configured maximum or when the
func (h *handler) manager() {
	defer close(h.doneCh) // indicate we're done, so Close() can return

	h.printDebug("manager started.")
	defer h.printDebug("manager terminated")

	ticker := time.NewTicker(h.cfg.MaxBatchDelay)
	batch := make([]NREvent, 0, h.cfg.MaxBatchSize)
	defer func() { h.postToInsights(batch) }() // make sure we write anything remaining in the batch to Insights

	for {
		h.printDebug("manager batch len %v.", len(batch))
		if len(batch) >= h.cfg.MaxBatchSize {
			go h.postToInsights(batch)
			batch = make([]NREvent, 0, h.cfg.MaxBatchSize)
		}

		select {
		case <-h.quitCh:
			h.printDebug("manager waiting for workers to finish.")
			h.wg.Wait() // wait for the workers to send their batches to new relic
			return
		case message := <-h.msgCh:
			batch = append(batch, message)
		case <-ticker.C:
			go h.postToInsights(batch)
			batch = make([]NREvent, 0, h.cfg.MaxBatchSize)
		}
	}
}

func (h *handler) postToInsights(s []NREvent) {
	// Make sure we don't get interrupted
	h.wg.Add(1)
	defer h.wg.Done()

	if len(s) == 0 {
		return
	}
	h.printDebug("posting %v events to Insights", len(s))

	url := h.cfg.Endpoint
	d, err := json.Marshal(s)
	if err != nil {
		log.Printf("Error marshalling batch for send to NR Insights: %v", err)
		h.requeue(s)
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(d))
	req.Header.Set("X-Insert-Key", h.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.requeue(s)
		log.Printf("Error sending batch to NR Insights: %v", err)
		return
	}
	h.printDebug("%v events successfully posted to Insights", len(s))
	defer resp.Body.Close()
}

func (h *handler) requeue(s []NREvent) {
	for _, e := range s {
		select {
		case <-h.quitCh: // if we're shutting down, abort the message sends.  The data will simply be lost
			return
		default:
			h.msgCh <- e
		}
	}
}

func (h *handler) printDebug(format string, a ...interface{}) {
	log.Debugf("[nrinsights] %v", fmt.Sprintf(format, a...))
}
